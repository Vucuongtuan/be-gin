package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	ID    string `json:"_id" bson:"_id"`
	Name  string `json:"name" bson:"name"`
	Email string `json:"email" bson:"email"`
	Exp   string `json:"exp" bson:"exp"`
	jwt.StandardClaims
}

func GetIdAuthorFromToken(c *gin.Context) (string, error) {
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {

		return "", errors.New("err")
	}
	parts := strings.Split(accessToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {

		return "", errors.New("err")
	}

	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(secretKey) == 0 {

		return "", errors.New("Can't get secret key")
	}

	token, err := jwt.ParseWithClaims(parts[1], &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {

		return "", err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {

		return "", errors.New("Can't claims token")
	}

	return claims.Id, nil
}
func GetIdAuthorFromTokenMidd(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"msg":    "Account unauthorized",
			"err":    "Can't get token from headers Authorization",
		})
		c.Abort()
		return
	}
	parts := strings.Split(accessToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {

		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"msg":    "Account unauthorized",
			"err":    "Can't parse Bearer Token",
		})
		c.Abort()
		return
	}

	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(secretKey) == 0 {

		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"msg":    "Account unauthorized",
			"err":    "Can't get secret key",
		})
		c.Abort()
		return
	}

	token, err := jwt.ParseWithClaims(parts[1], &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {

		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"msg":    "Account unauthorized",
			"err":    "Can't parse Token",
		})
		c.Abort()
		return
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"msg":    "Account unauthorized",
			"err":    "Can't claims Token",
		})
		c.Abort()
		return
	}
	c.Set("id", claims.ID)
	c.Next()
}
