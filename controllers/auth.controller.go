package controllers

import (
	"be/models"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
)

type DataToken struct {
	ID    string `json:"_id" bson:"_id"`
	Name  string `json:"name" bson:"_name"`
	email string `json:"email" bson:"email"`
}

func Login(c *gin.Context) {
	name, _ := c.Get("name")
	email, _ := c.Get("email")
	_id, _ := c.Get("_id")

	conn := models.NewConn()

	expToken := time.Now().Add(time.Hour * 24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":   expToken,
		"name":  name,
		"email": email,
		"_id":   _id,
	})
	var existingUser bson.M
	err := conn.CollectionOnline.FindOne(context.Background(), bson.M{"token": token}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"token":  existingUser["token"],
			"exp":    int(existingUser["exp"].(time.Time).Sub(time.Now()).Seconds()),
		})
		return
	}

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Failed to generate token",
			"err": err,
		})
		return
	}
	_, err = conn.CollectionOnline.InsertOne(context.Background(), bson.M{"token": tokenString, "exp": expToken})
	if err != nil {
		c.JSON(http.StatusNotExtended, gin.H{
			"status": http.StatusNotExtended,
			"msg":    "Can't login to database online",
			"err":    err,
		})
		return
	}
	maxAge := int(expToken.Sub(time.Now()).Seconds())
	c.SetCookie("access_token", tokenString, maxAge, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "Login successfully authenticated",
		"data": bson.M{
			"name":  name,
			"email": email,
			"_id":   _id,
		},
		"access_token": tokenString,
		"exp":          maxAge,
	})
}

type ReqResetPassword struct {
	Email           string     `json:"email" bson:"email"`
	Password        string     `json:"password" bson:"password"`
	ComfirmPassword string     `json:"confirm_password" bson:"confirm_password"`
	Otp             int64      `json:"otp" bson:"otp"`
	Created_At      *time.Time `json:"created_at" bson:"created_at"`
}

func ResetPassword(c *gin.Context) {
	var resetPasswordDto ReqResetPassword
	if err := c.ShouldBindJSON(&resetPasswordDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Can't bind data json request",
			"err":    err,
		})
		return
	}
	if resetPasswordDto.Password != resetPasswordDto.ComfirmPassword {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"msg":    "Password does not match comfirmPassword",
		})
		return
	}
	conn := models.NewConn()
	status, msg, err := models.ResetPassword(conn, resetPasswordDto.Otp, resetPasswordDto.Email, resetPasswordDto.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Can't reset password , please try again",
			"data":   resetPasswordDto,
			"err":    err,
			"a":      msg,
			"s":      status,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "Password reset successfully",
		"a":      msg,
		"s":      status,
	})
}

func LogoutController(c *gin.Context) {
	var token string

	if err := c.ShouldBindJSON(&token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Can't find token from request",
			"err":    err,
		})
		return
	}

	err := models.LogoutAccount(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Can't logout account",
			"err":    err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "Logout account successfully",
	})
}
