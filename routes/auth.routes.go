package routes

import (
	"be/controllers"
	"be/middleware"

	"github.com/gin-gonic/gin"
)

type A struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var FakeData = []A{
	{
		ID:      1,
		Title:   "Blog Post 1",
		Content: "This is the content of Blog Post 1.",
	},
	{
		ID:      2,
		Title:   "Blog Post 2",
		Content: "This is the content of Blog Post 2.",
	},
	{
		ID:      3,
		Title:   "Blog Post 3",
		Content: "This is the content of Blog Post 3.",
	},
}

func authRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/send-otp", middleware.SendMail)
		auth.POST("/register", controllers.CreateUser)
		auth.POST("/login", middleware.LoginMiddleware, controllers.Login)
		auth.PATCH("/reset-password", controllers.ResetPassword)
		auth.DELETE("/logout", controllers.LogoutController)
	}
}
