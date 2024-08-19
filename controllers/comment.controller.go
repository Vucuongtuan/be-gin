package controllers

import (
	"be/models"
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

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var clients = make(map[string][]*websocket.Conn)
var mutex = sync.Mutex{}

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

func SocketComment(c *gin.Context, blogID string, UserID string) {

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
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

		_, _, _ = model.CommentByBlog(blogOBID, userID, commentData.Content)

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
func SocketLikeAndDisLikeComment(c *gin.Context) {
	commentID := c.Param("commentID")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing commentID"})
		return
	}

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
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

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
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
		_, _, _ = model.CommentByBlog(msg.BlogID, msg.UserID, msg.Content)

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

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
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

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
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
		_, action, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, blogID)
			mutex.Unlock()
			return
		}

		var data ActionLikeOrDisLike
		if err := json.Unmarshal(action, &data); err != nil {
			continue
		}

		model := models.NewConn()
		if data.Type == "like" {
			_ = model.LikeBlog(data.UserID, data.BlogID)
		} else {
			_ = model.DisLikeBlog(data.UserID, data.BlogID)
		}

		mutex.Lock()
		for _, client := range clients[blogID] {
			if err := client.WriteJSON(data); err != nil {
				client.Close()
				delete(clients, blogID)
			}
		}
		mutex.Unlock()
	}
}

// func wshandler(w http.ResponseWriter, r *http.Request) {
//     conn, err := wsupgrader.Upgrade(w, r, nil)
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
