package controllers

import (
	"be/middleware"
	"be/models"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type createBlogsDto struct {
	Title       string   `bson:"title" json:"title"`
	Description string   `bson:"description" json:"description"`
	Slug        *string  `bson:"slug" json:"slug"`
	Hashtags    []string `bson:"hashtags" json:"hashtags"`
	Link        *string  `bson:"link" json:"link"`
	Type        *string  `bson:"type" json:"type"`
	Action      string   `bson:"action" json:"action"`
}

func GetAllBlogs(c *gin.Context) {
	// get query parameters from url api
	var req struct {
		Page  int64 `form:"page"`
		Limit int64 `form:"limit"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid query parameters",
			"error":  err.Error(),
		})
		return
	}
	// Set default values for page and limit
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit, _ = strconv.ParseInt(os.Getenv("LIMIT"), 10, 64)
	}
	// create connection func NewConn in models
	conn := models.NewConn()

	blogs, total, totalPages, err := conn.GetAllBlogs(req.Page, req.Limit)
	///check err
	if err != nil {
		//return response error
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Can't get all blog",
			"err":    err.Error(),
		})
		return
	}
	if total == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"msg":    "Blogs not found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "Get all blog successfully",
		"current": gin.H{
			"total":      total,
			"totalPages": totalPages,
			"page":       req.Page,
		},
		"data": blogs,
	})
}

func GetBlogNewFeatured(c *gin.Context) {
	var req struct {
		Page  int64 `form:"page"`
		Limit int64 `form:"limit"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid query parameters",
			"error":  err.Error(),
		})
		return
	}
	// Set default values for page and limit
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit, _ = strconv.ParseInt(os.Getenv("LIMIT"), 10, 64)
	}

	// create connection func NewConn in models
	conn := models.NewConn()

	blogs, total, totalPages, err := conn.GetBlogNewFeatured(req.Page, req.Limit)
	///check err
	if err != nil {
		//return response error
		c.JSON(http.StatusAlreadyReported, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Can't get all users",
			"err":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "Get all blog successfully",
		"current": gin.H{
			"total":      total,
			"totalPages": totalPages,
			"page":       req.Page,
		},
		"data": blogs,
	})
}
func GetBlogDetailBySlug(c *gin.Context) {
	var user struct {
		idUser string `bson:"id_user" json:"id_user"`
	}

	if err := c.ShouldBindQuery(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid query parameters",
			"error":  err.Error(),
		})
		return
	}

	slug := c.Query("slug")
	if slug == "" {
		c.JSON(http.StatusNoContent, gin.H{
			"status": http.StatusNoContent,
			"msg":    "Can't get slug from query parameter",
		})
		return
	}
	model := models.NewConn()
	status, msg, blog, err := model.GetBlogDetailBySlug(slug)
	if err != nil {
		c.JSON(status, gin.H{
			"status": status,
			"msg":    msg,
			"err":    err,
		})
		return
	}
	c.JSON(status, gin.H{
		"status": status,
		"msg":    msg,
		"data":   blog,
	})
}
func GetBlogByAuthor(c *gin.Context) {
	idAuthor := c.Param("id")
	pageStr := c.Query("page")
	if idAuthor == "" {
		c.JSON(http.StatusNoContent, gin.H{
			"status": http.StatusNoContent,
			"msg":    "Author not found",
		})
	}
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid page number",
		})
		return
	}
	idAuthorObj, _ := primitive.ObjectIDFromHex(idAuthor)
	conn := models.NewConn()
	blogs, status, msg, err := conn.GetBlogByAuthor(idAuthorObj, page)
	if err != nil {
		c.JSON(status, gin.H{
			"status": status,
			"msg":    msg,
			"err":    err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"msg":    msg,
		"data":   blogs,
	})
}
func CreateBlog(c *gin.Context) {
	idAuthor := c.Param("id")
	if idAuthor == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":    "Can't get id author",
			"status": http.StatusBadRequest,
		})
		return

	}
	idAuthorObj, _ := primitive.ObjectIDFromHex(idAuthor)
	var createBlogDto models.CreateBlogsDto
	if err := c.ShouldBind(&createBlogDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": 400,
			"msg":    "Failed to parse form",
		})
		return
	}
	valiToken, err := middleware.GetIdAuthorFromToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
		return
	}

	conn := models.NewConn()
	status, msg, err := conn.CreateBlog(createBlogDto, idAuthorObj, c)

	if err != nil {
		c.JSON(status, gin.H{
			"status": status,
			"msg":    msg,
			"err":    err,
		})
		return
	}

	c.JSON(status, gin.H{
		"status": status,
		"msg":    msg,
		"data":   createBlogDto,
		"id":     valiToken,
	})

}

func UpdateBlog(c *gin.Context) {
	id := c.Param("id")
	var updateBlogDto models.Blogs

	conn := models.NewConn()
	status, msg, err := conn.UpdateBlog(id, updateBlogDto)
	if err != nil {
		c.JSON(status, gin.H{
			"status": status,
			"msg":    msg,
			"err":    err,
		})
		return
	}
	c.JSON(status, gin.H{
		"status": status,
		"msg":    msg,
	})
}
func DeleteBlog(c *gin.Context) {
	id := c.Param("id")
	conn := models.NewConn()
	status, msg, err := conn.DeleteBlog(id)
	if err != nil {
		c.JSON(status, gin.H{
			"status": status,
			"msg":    msg,
			"err":    err,
		})
		return
	}
	c.JSON(status, gin.H{
		"status": status,
		"msg":    msg,
	})
}
func GetBlogUserLike(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusNoContent, gin.H{
			"status": http.StatusNoContent,
			"msg":    "User not found",
		})
		return
	}
	model := models.NewConn()
	idObj, _ := primitive.ObjectIDFromHex(id)
	data, err := model.GetBlogUserLike(idObj)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNoContent, gin.H{
				"status": http.StatusNoContent,
				"msg":    "Data not found",
				"err":    err,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Failed to get data user like",
			"err":    err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "OK",
		"data":   data,
	})
}
