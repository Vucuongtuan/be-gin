package routes

import (
	"be/controllers"
	"be/middleware"
	"be/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ,func (c *gin.Context){
// 	c.JSON(http.StatusOK,gin.H{
// 		"status":http.StatusOK,
// 		"msg":http.StatusOK,
// 	})
// }
func routesBlogs(r *gin.RouterGroup) {
	blogs := r.Group("/blogs")
	{
		blogs.GET("/all",middleware.Authoriation())
		blogs.GET("/new",controllers.GetBlogNewFeatured)
		blogs.GET("/", middleware.Authoriation(), controllers.GetAllBlogs)
		blogs.GET("/q",  controllers.GetBlogDetailBySlug)
		blogs.POST("/", middleware.UploadFile, controllers.CreateBlog)
		blogs.PATCH("/:id", middleware.UploadFile, controllers.UpdateBlog)
		blogs.DELETE("/:id", controllers.DeleteBlog)
		blogs.POST("/test",func (c *gin.Context) {
			var createBlogDto models.CreateBlogsDto

			if err := c.ShouldBind(&createBlogDto); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		
			// title := c.PostForm("title")
			c.JSON(http.StatusOK, gin.H{
				"data": createBlogDto,
				"time":time.Now().UTC(),
			})

		})
	}
}