package routes

import (
	"be/controllers"
	"be/middleware"
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
			userID, err := middleware.GetIdAuthorFromToken(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "msg": "Can't comment by blog", "status": http.StatusInternalServerError})
				return
			}
			controllers.SocketComment(c, blogID, userID)
		})
		ws.GET("/reply/:blogID", controllers.SocketReplyComment)
		ws.GET("/rec-blog/:blogID", controllers.SocketLikeAndDisLikeBlog)
		ws.GET("/rec-comment/:commentID", controllers.SocketLikeAndDisLikeComment)
		ws.GET("/rec-reply/:commentID", controllers.SocketLikeOrDislikeReply)
		ws.GET("/notifications/all", controllers.GetAllNotificationsByUserID)
		ws.GET("/notifications", controllers.GetAllNotificationsByUserID)
		ws.POST("/follow", middleware.GetIdAuthorFromTokenMidd, controllers.Follow)

	}
}
