package routes

import (
	"be/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func socketRoutes(r *gin.RouterGroup) {

	ws := r.Group("/ws")
	{
		ws.GET("/:blogID", func(c *gin.Context) {
			blogID := c.Param("blogID")
			if blogID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Missing blog_id"})
				return
			}

			controllers.SocketComment(c, blogID)
		})
		ws.GET("/reply/:blogID", controllers.SocketReplyComment)
		ws.GET("/rec-blog/:blogID", controllers.SocketLikeAndDisLikeBlog)
		ws.GET("/rec-comment/:commentID", controllers.SocketLikeAndDisLikeComment)
		ws.GET("/rec-reply/:commentID", controllers.SocketLikeOrDislikeReply)
		// ws.GET("/like",controll)
	}
}
