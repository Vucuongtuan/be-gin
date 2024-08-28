package routes

import (
	"be/controllers"
	"be/graphQL"
	"be/middleware"
	"be/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

//	,func (c *gin.Context){
//		c.JSON(http.StatusOK,gin.H{
//			"status":http.StatusOK,
//			"msg":http.StatusOK,
//		})
//	}

type SuggestRequest struct {
	Keyword string `json:"keyword" bson:"keyword"`
}

func routesBlogs(r *gin.RouterGroup) {
	blogs := r.Group("/blogs")
	{
		blogs.GET("/all", middleware.Authoriation())
		blogs.GET("/new", controllers.GetBlogNewFeatured)
		blogs.GET("/", controllers.GetAllBlogs)
		blogs.GET("/author/:id", controllers.GetBlogByAuthor)
		blogs.GET("/q", middleware.View, controllers.GetBlogDetailBySlug)
		blogs.POST("/search", func(c *gin.Context) {
			var req SuggestRequest
			if err := c.BindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"err":    err.Error(),
					"msg":    "Invalid request body",
					"status": http.StatusBadRequest,
				})
				return
			}

			conn := models.NewConn()
			blogsData, err := conn.SearchBlogsByHashtag(req.Keyword)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"err":    err.Error(),
					"msg":    "Failed to get blogs",
					"status": http.StatusInternalServerError,
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusOK,
				"msg":    "OK",
				"data":   blogsData,
			})
		})
		blogs.POST("/suggest", func(c *gin.Context) {
			var req SuggestRequest
			if err := c.BindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"err":    err.Error(),
					"msg":    "Invalid request body",
					"status": http.StatusBadRequest,
				})
				return
			}

			conn := models.NewConn()
			relatedHashtags, err := conn.GetRelatedHashtags(req.Keyword)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"err":    err.Error(),
					"msg":    "Failed to get related hashtags",
					"status": http.StatusInternalServerError,
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusOK,
				"msg":    "OK",
				"data":   relatedHashtags,
			})
		})
		blogs.POST("/author/:id", middleware.UploadFile, controllers.CreateBlog)

		blogs.POST("/rec-blog", graphQL.ActionLikeOrDislike)
		blogs.PATCH("/:id", middleware.UploadFile, controllers.UpdateBlog)
		blogs.DELETE("/:id", controllers.DeleteBlog)
		blogs.POST("/test", func(c *gin.Context) {
			var createBlogDto models.CreateBlogsDto

			if err := c.ShouldBind(&createBlogDto); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// title := c.PostForm("title")
			c.JSON(http.StatusOK, gin.H{
				"data": createBlogDto,
				"time": time.Now().UTC(),
			})

		})
	}
}
