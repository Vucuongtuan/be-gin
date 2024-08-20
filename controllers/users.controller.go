package controllers

import (
	"be/helpers"
	"be/models"
	"be/socket"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// func get all users
func GetAllUsers(c *gin.Context) {
	// get query parameters from url api
	pageQuery := c.DefaultQuery("page", "1")
	limitQuery := c.DefaultQuery("limit", os.Getenv("LIMIT"))

	page, _ := strconv.ParseInt(pageQuery, 10, 64)
	limit, _ := strconv.ParseInt(limitQuery, 10, 64)
	// create connection func NewConn in models
	conn := models.NewConn()

	// pass page and limit into the function GetAllUsers
	users, total, totalPages, err := conn.GetAllUsers(page, limit)

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
		"status":     http.StatusOK,
		"msg":        "Get all users successfully",
		"total":      total,
		"totalPages": totalPages,
		"page":       page,
		"data":       users,
	})
}
func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	idObj, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid user ID",
			"err":    err.Error(),
		})
		return
	}

	conn := models.NewConn()

	users, err := conn.GetUserByID(idObj)
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
		"msg":    "Get all users successfully",
		"data":   users,
	})
}

// func register create user end account
func CreateUser(c *gin.Context) {
	var createUserDto models.CreateUser
	if err := c.ShouldBindJSON(&createUserDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"data": createUserDto,
		})
		return
	}
	time := time.Now()
	conn := models.NewConn()
	account := models.Account{
		Name_Account: createUserDto.NameAccount,
		Password:     createUserDto.Password,
		Created_At:   &time,
	}
	create_account, err := conn.CreateAccount(account)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{

			"status": http.StatusBadRequest,
			"msg":    "Failed to create account , please try again later",
			"err":    err,
		})
		return
	}

	user := models.User{
		Name:          createUserDto.Name,
		Email:         createUserDto.Email,
		Date_BirthDay: createUserDto.DateBirth,
		Account:       create_account.Hex(),

		Created_At: &time,
	}

	// create_user,err :=
	err = conn.CreateUser(user)
	if err != nil {
		err = conn.DeleteUser(create_account)
		c.JSON(http.StatusBadRequest, gin.H{

			"status": http.StatusBadRequest,
			"msg":    "Failed to create account,please try again",
			"err":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"msg":    "Created successfully",
		"status": http.StatusCreated,
		// "conn":conn,
	})
	return

}

func DeleteAccountEndUser(c *gin.Context) {
	_id := c.Param("id")
	conn := models.NewConn()
	objectId, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":    "Invalid user ID",
			"status": http.StatusBadRequest,
		})
		return
	}

	err = conn.DeleteUser(objectId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":    "Can not delete user from database",
			"status": http.StatusBadRequest,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "User deleted successfully",
	})
	return

}

func UpdateUser(c *gin.Context) {
	var updateUserDTO models.UPdateUser
	_id := c.Param("id")
	err := c.ShouldBindJSON(&updateUserDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":    "Invalid request body",
			"status": http.StatusBadRequest,
			"err":    err.Error(),
		})
		return
	}
	conn := models.NewConn()
	objectId, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":    "Invalid user ID",
			"status": http.StatusBadRequest,
			"err":    err.Error(),
		})
		return
	}
	err = conn.UpdateUser(objectId, updateUserDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":    "Can not update user from database",
			"status": http.StatusBadRequest,
			"err":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "User update successfully",
	})
	return

}

func Follow(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Failed to upgrade to WebSocket connection",
			"err":    err.Error(),
		})
		return
	}
	defer conn.Close()

	var userFollow struct {
		ID       string `json:"_id" bson:"_id"`
		IDFollow string `json:"id_follow" bson:"id_follow"`
		Name     string `json:"name" bson:"name"`
	}
	if err := c.ShouldBindJSON(&userFollow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid request body",
			"err":    err.Error(),
		})
		return
	}

	err = conn.ReadJSON(&userFollow)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid request body",
			"err":    err.Error(),
		})
		return
	}
	model := models.NewConn()
	err = model.Follow(userFollow.ID, userFollow.IDFollow, userFollow.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Can not follow user",
			"err":    err.Error(),
		})
		return
	}
	notification := socket.Notification{
		FromUserID: userFollow.ID,
		ToUserID:   userFollow.IDFollow,
		Message:    userFollow.Name + " đã theo dõi bạn.",
	}
	if err := helpers.SendNotification(notification.ToUserID, notification.FromUserID, notification.Message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Failed to send notification",
			"err":    err.Error(),
		})
		return
	}
	go func() {
		socket.Broadcast <- notification
	}()

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "User follow successfully",
	})
}
func UnFollow(c *gin.Context) {
	var userUnFollow struct {
		ID       string `json:"_id" bson:"_id"`
		IDFollow string `json:"id_follow" bson:"id_follow"`
	}
	if err := c.ShouldBindJSON(&userUnFollow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid request body",
			"err":    err.Error(),
		})
		return
	}
	model := models.NewConn()
	err := model.UnFollow(userUnFollow.ID, userUnFollow.IDFollow)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Can not unfollow user",
			"err":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "User unfollow successfully",
	})
}
func GetRecentNotificationsByUserID(c *gin.Context) {
	id := c.Param("id")
	conn := models.NewConn()
	idObj, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid user ID",
			"err":    err.Error(),
		})
		return
	}
	nof, err := conn.GetRecentNotificationsByUserID(idObj)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Can not get recent notifications",
			"err":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "Recent notifications successfully",
		"data":   nof,
	})
}
func GetAllNotificationsByUserID(c *gin.Context) {
	id := c.Param("id")
	conn := models.NewConn()
	idObj, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid user ID",
			"err":    err,
		})
		return
	}
	nof, err := conn.GetAllNotificationsByUserID(idObj)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Invalid user ID",
			"err":    err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "Get OK",
		"data":   nof,
	})
}
