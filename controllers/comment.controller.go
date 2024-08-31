package controllers

import (
	"be/helpers"
	"be/models"
	"be/socket"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentByBlogDTO struct {
	BlogID  string `bson:"blog_id" json:"blog_id"`
	Content string `bson:"content" json:"content"`
	UserID  string `bson:"user_id" json:"user_id"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var clients = make(map[string][]*websocket.Conn)
var mutex = sync.Mutex{}

type BlogCommentClients struct {
	Clients map[string]*websocket.Conn
	Mutex   sync.Mutex
}

// var blogCommentClients = make(map[string]*BlogCommentClients)
var blogCommentClientsMutex = sync.Mutex{}
var blogCommentClients = &BlogCommentClients{
	Clients: make(map[string]*websocket.Conn),
	Mutex:   sync.Mutex{},
}

type Message struct {
	Username   string              `json:"username"`
	BlogID     primitive.ObjectID  `json:"blog_id" bson:"blog_id"`
	UserID     primitive.ObjectID  `json:"user_id" bson:"user_id"`
	Content    string              `json:"content" bson:"content"`
	CommentID  *primitive.ObjectID `json:"comment_id" bson:"comment_id"`
	Created_At *time.Time          `json:"created_at" bson:"created_at"`
}
type ActionLikeOrDisLike struct {
	Type       string              `json:"type" bson:"type"`
	UserID     primitive.ObjectID  `json:"user_id" bson:"user_id"`
	CommentID  *primitive.ObjectID `json:"comment_id" bson:"comment_id"`
	BlogID     primitive.ObjectID  `json:"blog_id" bson:"blog_id"`
	Created_At *time.Time          `json:"created_at" bson:"created_at"`
}

func CommentByBlog(c *gin.Context) {
	id := c.Param("blogID")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Can't get id blog by params",
		})
	}
	model := models.NewConn()
	idObj, _ := primitive.ObjectIDFromHex(id)
	blog, err := model.GetCommentByBlog(idObj)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"data":   blog,
		"msg":    "Get blog OK",
	})
}
func SocketComment(c *gin.Context, blogID string, UserID string) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	mutex.Lock()
	clients[blogID] = append(clients[blogID], conn)
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		connections := clients[blogID]
		for i, c := range connections {
			if c == conn {
				clients[blogID] = append(connections[:i], connections[i+1:]...)
				break
			}
		}
		mutex.Unlock()
	}()
	for {

		var commentData struct {
			Content string `json:"content"`
		}
		err := conn.ReadJSON(&commentData)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Error(err)
			}
			break
		}
		blogOBID, _ := primitive.ObjectIDFromHex(blogID)
		userID, _ := primitive.ObjectIDFromHex(UserID)

		model := models.NewConn()

		_, _ = model.CommentByBlog(blogOBID, userID, commentData.Content)

		user, err := model.GetUserByID(userID)
		mutex.Lock()

		for _, c := range clients[blogID] {
			if c != conn {
				c.WriteJSON(gin.H{
					"status":  http.StatusOK,
					"message": "New comment added",
					"data": gin.H{
						"user_id":  userID,
						"avatar":   user.Avatar,
						"name":     user.Name,
						"message":  commentData.Content,
						"blog_id":  blogID,
						"datetime": time.Now(),
					},
				})
			}
		}
		mutex.Unlock()
	}
}

func CommmentByBlog(c *gin.Context) {
	var dataReq helpers.ReqDataComment
	if err := c.BindJSON(&dataReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Err,reqComment",
			"err":    err,
		})
		return
	}
	model := models.NewConn()
	idBlogObj, _ := primitive.ObjectIDFromHex(dataReq.IDBlog)
	idUserObj, _ := primitive.ObjectIDFromHex(dataReq.UserID)

	_, err := model.CommentByBlog(idBlogObj, idUserObj, dataReq.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Err",
			"err":    err.Error(),
		})
		return
	}
	if err := helpers.SendCommentAllUser(dataReq.IDBlog, dataReq.UserID, dataReq.Name, dataReq.Avatar, dataReq.Message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Failed to send notification",
			"err":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "OK",
		"data":   dataReq,
	})
}
func SocketLikeAndDisLikeComment(c *gin.Context) {
	commentID := c.Param("commentID")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing commentID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect websocket"})
		return
	}
	defer conn.Close()

	mutex.Lock()
	if _, exists := clients[commentID]; !exists {
		clients[commentID] = []*websocket.Conn{}
	}
	clients[commentID] = append(clients[commentID], conn)
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		conns := clients[commentID]
		for i, c := range conns {
			if c == conn {
				clients[commentID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		mutex.Unlock()
	}()

	for {
		_, action, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, commentID)
			mutex.Unlock()
			return
		}

		var data ActionLikeOrDisLike
		if err := json.Unmarshal(action, &data); err != nil {
			continue
		}

		model := models.NewConn()
		if data.Type == "like" {
			_ = model.CommentLike(data.UserID, *data.CommentID)
		} else {
			_ = model.CommentDisLike(data.UserID, *data.CommentID)
		}

		mutex.Lock()
		for _, client := range clients[commentID] {
			if err := client.WriteJSON(data); err != nil {
				client.Close()
				delete(clients, commentID)
			}
		}
		mutex.Unlock()
	}
}

func SocketReplyComment(c *gin.Context) {
	blogID := c.Param("blogID")
	if blogID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing blogID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect websocket"})
		return
	}
	defer conn.Close()

	mutex.Lock()
	if _, exists := clients[blogID]; !exists {
		clients[blogID] = []*websocket.Conn{}
	}
	clients[blogID] = append(clients[blogID], conn)
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		conns := clients[blogID]
		for i, c := range conns {
			if c == conn {
				clients[blogID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		mutex.Unlock()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, blogID)
			mutex.Unlock()
			return
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		model := models.NewConn()
		_, _ = model.CommentByBlog(msg.BlogID, msg.UserID, msg.Content)

		mutex.Lock()
		for _, client := range clients[blogID] {
			if err := client.WriteJSON(msg); err != nil {
				client.Close()
				delete(clients, blogID)
			}
		}
		mutex.Unlock()
	}
}

func SocketLikeOrDislikeReply(c *gin.Context) {
	commentID := c.Param("commentID")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing commentID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect websocket"})
		return
	}
	defer conn.Close()

	mutex.Lock()
	if _, exists := clients[commentID]; !exists {
		clients[commentID] = []*websocket.Conn{}
	}
	clients[commentID] = append(clients[commentID], conn)
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		conns := clients[commentID]
		for i, c := range conns {
			if c == conn {
				clients[commentID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		mutex.Unlock()
	}()

	for {
		_, action, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, commentID)
			mutex.Unlock()
			return
		}

		var data ActionLikeOrDisLike
		if err := json.Unmarshal(action, &data); err != nil {
			continue
		}

		model := models.NewConn()
		if data.Type == "like" {
			_ = model.ReplyCommentLike(data.UserID, *data.CommentID)
		} else {
			_ = model.ReplyCommentDisLike(data.UserID, *data.CommentID)
		}

		mutex.Lock()
		for _, client := range clients[commentID] {
			if err := client.WriteJSON(data); err != nil {
				client.Close()
				delete(clients, commentID)
			}
		}
		mutex.Unlock()
	}
}

func SocketLikeAndDisLikeBlog(c *gin.Context) {
	blogID := c.Param("blogID")
	if blogID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing blogID"})
		return
	}

	var data ActionLikeOrDisLike
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error(),
			"msg":    "Can't reaction blog",
			"status": http.StatusBadRequest,
		})
		return
	}

	model := models.NewConn()
	if data.Type == "like" {
		err := model.LikeBlog(data.UserID, data.BlogID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error(), "msg": "Can't like blog", "status": http.StatusInternalServerError})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"msg":    "ok",
		})
		return
	} else {
		err := model.DisLikeBlog(data.UserID, data.BlogID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error(), "msg": "Can't like blog", "status": http.StatusInternalServerError})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"msg":    "ok",
		})
		return
	}

}
func NotifyWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Failed to upgrade to WebSocket connection",
			"err":    err.Error(),
		})
		return
	}
	defer conn.Close()

	// Add the connection to the notification hub

	// Listen for notifications and send them to the client
	for notification := range socket.Broadcast {
		err = conn.WriteJSON(notification)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    "Failed to send WebSocket message",
				"err":    err.Error(),
			})
			return
		}
	}
}

// func wshandler(w http.ResponseWriter, r *http.Request) {
//     conn, err := upgrader.Upgrade(w, r, nil)
//     if err != nil {
//         fmt.Println("Failed to set websocket upgrade: %+v", err)
//         return
//     }

//     for {
//         t, msg, err := conn.ReadMessage()
//         if err != nil {
//             break
//         }
//         conn.WriteMessage(t, msg)
//     }
// }
