package middleware

import (
	"be/helpers"
	"be/models"
	"be/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SendMail(c *gin.Context) {
	var EmailDto struct {
		Email string `json:"email" bson:"email"`
	}
	if err := c.ShouldBindJSON(&EmailDto); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
					"status": http.StatusNotFound,
					"msg":    "Can't get email from request",
				})
				return
	}

	otp := utils.GenerateOtp()
	m:= models.NewConn()
	if err := m.AddOtp(int64(otp), EmailDto.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Can't add otp to database",
			"err":    err.Error(),
		})
		return
	}
	if err := helpers.SendEmail(EmailDto.Email, otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Can't send otp to " + EmailDto.Email,
			"err":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "Send mail successfully",
	})
}
