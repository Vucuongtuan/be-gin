package middleware

import (
	"be/models"
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func  View(blogID primitive.ObjectID ,c *gin.Context)  {

	conn := models.NewConn()
	filter := bson.M{
		"$inc": bson.M{
			"view": 1,
		},
	}
	_, err := conn.CollectionBlogs.UpdateByID(context.Background(), blogID, filter)
	if err != nil {
		c.Next()
	}
	c.Next()
}