package routes

import (
	"be/controllers"
	"be/helpers"
	"be/middleware"
	"github.com/gin-gonic/gin"
)

func socketRoutes(r *gin.RouterGroup) {

	ws := r.Group("/ws")
	{
		ws.GET("/:blogID", helpers.SOcketCOmmentBlog)
		ws.GET("/reply/:blogID", controllers.SocketReplyComment)
		ws.POST("/rec-blog/:blogID", controllers.SocketLikeAndDisLikeBlog)
		ws.GET("/rec-comment/:commentID", controllers.SocketLikeAndDisLikeComment)
		ws.GET("/rec-reply/:commentID", controllers.SocketLikeOrDislikeReply)
		ws.GET("/notifications/all", controllers.GetAllNotificationsByUserID)
		ws.GET("/notifications/:id", helpers.WebSocketHandler)
		ws.POST("/follow", middleware.GetIdAuthorFromTokenMidd, controllers.Follow)

	}
}
