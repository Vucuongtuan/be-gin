package middleware

import (
	"errors"
	"fmt"
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
func GetIdAuthorFromToken(c *gin.Context) (Claims, error) {
    var emptyClaims Claims 

    accessToken := c.GetHeader("Authorization")
    if accessToken == "" {
        return emptyClaims, errors.New("authorization header missing")
    }

    parts := strings.Split(accessToken, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return emptyClaims, errors.New("invalid token format")
    }

    secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
    if len(secretKey) == 0 {
        return emptyClaims, errors.New("secret key not found")
    }

    token, err := jwt.ParseWithClaims(parts[1], &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return secretKey, nil
    })

    if err != nil || !token.Valid {
        return emptyClaims, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok {
        return emptyClaims, errors.New("unable to parse claims or ID missing")
    }
    return *claims, nil
}
