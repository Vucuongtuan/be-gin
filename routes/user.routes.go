package routes

import (
	"be/controllers"

	"github.com/gin-gonic/gin"
)

func userRoutes(r *gin.RouterGroup) {

	users := r.Group("/users")
	{
		users.GET("/", controllers.GetAllUsers)
		users.POST("/create", controllers.CreateUser)
		users.POST("/un-follow", controllers.UnFollow)
		users.GET("/ws/follow", controllers.Follow)
		users.DELETE("/:id", controllers.DeleteAccountEndUser)
		users.PUT("/:id", controllers.UpdateUser)

	}
}

// users.PUT("/update", controllers.updateUser)
// users.DELETE("/delete", controllers.deleteUser)
