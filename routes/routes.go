package routes

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.RouterGroup) {
	// user router => user.routes.go
	userRoutes(r)
	//auth router => auth.routes.go
	authRoutes(r)
	
	// blog router => blog.routes.go
	routesBlogs(r)
	
	socketRoutes(r)
}
