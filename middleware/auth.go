package middleware

import (
	"be/config"
	"be/models"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func middConn() models.Conn {
	return models.Conn{
		CollectionUser:    config.GetCollection("users"),
		CollectionAccount: config.GetCollection("accounts"),
		CollectionOnline:  config.GetCollection("online"),
	}
}

func LoginMiddleware(c *gin.Context) {
	conn := middConn()
	var data models.Login
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Don't handle data user",
			"err":    err.Error(),
		})
		c.Abort()
		return
	}

	dbAccount, err := models.CheckAccountLogin(data, conn, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Don't login account ,please try again",
			"err":    err.Error(),
		})
		c.Abort()
		return
	}
	var users struct {
		ID    primitive.ObjectID `bson:"_id" json:"_id"`
		Name  string             `bson:"name" json:"name"`
		Email string             `bson:"email" json:"email"`
	}
	filter := bson.M{
		"account": dbAccount.ID.Hex(),
	}
	err = conn.CollectionUser.FindOne(context.Background(), filter, options.FindOne().SetProjection(bson.M{
		"_id":   1,
		"name":  1,
		"email": 1,
	})).Decode(&users)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Don't handle data user",
			"err":    err.Error(),
			// "data":uss,
		})
		c.Abort()
		return
	}
    _ ,err = CheckEmail(users.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "The user is logged in elsewhere",
			"err":    err.Error(),
		})
		c.Abort()
        return
	}

	
	c.Set("name",users.Name)
	c.Set("email",users.Email)
	c.Set("_id",users.ID)
	c.Next()
}

// TUser,err := helpers.AuthUser(dbAccount)
// // err := conn.CollectionUser.FindOne(c, bson.M{"_id": dbAccount._id}).Decode(&TUser)
// if err != nil {
// 	c.AbortWithStatus(http.StatusUnauthorized, gin.H{
// 		"Status": http.StatusUnauthorized,
// 		"msg":    "Can't login account ,beacuse don't get user",
// 		"err":    err.Error(),
// 	})
// 	return
// }
// c.Set("TUser",bson.M{
// 	"_id":TUser._id,
// 	"name":TUser.Name,
// 	"email":TUser.Email,
// })
// c.Next()

func Authoriation() gin.HandlerFunc {
	return func(c *gin.Context) {

		accessToken := c.GetHeader("Authorization")
		fmt.Println(accessToken)
		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				            "status": http.StatusUnauthorized,
				            "msg":    "Can't get access token from request headers",
						
				        })
				        c.Abort()
            return
        }
		// if err != nil {
		// 	if err == http.ErrNoCookie {
		// 		c.Header("WWW-Authenticate", "Bearer")
		//         c.JSON(http.StatusUnauthorized, gin.H{
		//             "status": http.StatusUnauthorized,
		//             "msg":    "Access token not found",
		// 			"err":err,
		//         })
		//         c.Abort()
		//         return
		//     }
		// 	c.JSON(http.StatusFound, gin.H{
		// 		"status": http.StatusFound,
		// 		"msg":    "Access token unauthorized",
		// 		"err":    err,
		// 	})
		// 	  c.Abort()
		// 	return
		// }
		parts := strings.Split(accessToken, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": http.StatusUnauthorized,
				"msg":    "Access token format is invalid",
			})
			c.Abort()
			return
		}
		accessToken = parts[1]
		auth, err := ValidateToken(accessToken)
		if err != nil || auth == false {
			c.JSON(http.StatusFound, gin.H{
				"status": http.StatusFound,
				"msg":    "Access token unauthorized",
				"err":    err,
			})
			c.Abort()
			return
		}
	
		// if time.Now().Unix() > claims.exp {
		// 	c.JSON(http.StatusFound, gin.H{
		// 		"status": http.StatusFound,
		// 		"msg":    "Access token unauthorized",
		// 		"err":    err,
		// 	})
		// 	return
		// }
		// if !checkTokenInMongoDB(token, claims._id) {
		// 	c.JSON(400, gin.H{
		// 		"error": "Token not found in MongoDB",
		// 	})
		// 	return
		// }
		c.Next()
	}
}
func checkTokenInMongoDB(tokenString string, id primitive.ObjectID) bool {
	conn := models.NewConn()
	filter := bson.M{"token": tokenString}
	count, err := conn.CollectionOnline.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false
	}
	return count > 0
}
func ValidateToken(access_token string) (bool, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(secretKey) == 0 {
		return false, fmt.Errorf("JWT secret key is not set")
	}
	conn := models.NewConn()
	filter := bson.M{"token": access_token}
	_, err := conn.CollectionOnline.CountDocuments(context.Background(), filter)
	if err != nil {
        return false, err
    }
	return true, nil
}

func CheckEmail (email string) (bool ,error) {
	conn := models.NewConn()
	filter := bson.M{
		"email":email,
	}
	count ,err := conn.CollectionOnline.CountDocuments(context.Background(), filter)
	if err != nil {
		return false ,err
	}
	return count > 0 ,nil
}