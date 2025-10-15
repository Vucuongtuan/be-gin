package middleware

import (
	"be/models"
	"github.com/gin-gonic/gin"
)

func View(c *gin.Context) {
	slug := c.Query("slug")
	conn := models.NewConn()
	err := conn.View(slug)
	if err != nil {
		c.Abort()
		return
	}
	c.Next()
}
