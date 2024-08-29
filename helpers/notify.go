package helpers

import (
	"be/models"
	"be/socket"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"sync"
	"time"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var notifyClients = make(map[string]*websocket.Conn)
var notifyMutex = sync.Mutex{}

type BlogCommentClients struct {
	Clients map[string]*websocket.Conn
	Mutex   sync.Mutex
}

var blogCommentClients = make(map[string]*BlogCommentClients)
var blogCommentClientsMutex = sync.Mutex{}

func NotifyFollower(conn *websocket.Conn, from string, to string, message string, time time.Time) error {
	conn, err := getWebSocketConnection(to)
	if err != nil {
		return err
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return err
	}
	return nil
}
func getWebSocketConnection(id string) (*websocket.Conn, error) {
	// Implement logic to get the WebSocket connection for the given user ID
	// This could involve maintaining a map of user IDs to WebSocket connections
	// or using a WebSocket hub/registry to manage the connections
	// For simplicity, we'll just return a new connection for now
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:8080/users/ws?id=%s", id), nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
func WebSocketHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Missing id"})
		return
	}

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upgrade connection"})
		return
	}
	defer conn.Close()

	notifyMutex.Lock()
	notifyClients[id] = conn
	notifyMutex.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			notifyMutex.Lock()
			delete(notifyClients, id)
			notifyMutex.Unlock()
			break
		}

		if string(message) == "get_notifications" {
			idObj, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Invalid user ID"}`))
				continue
			}

			dbConn := models.NewConn()
			notifications, err := dbConn.GetAllNotificationsByUserID(idObj)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Failed to fetch notifications"}`))
				continue
			}

			notificationsJSON, _ := json.Marshal(notifications)
			conn.WriteMessage(websocket.TextMessage, notificationsJSON)
		}
	}
}

func SendNotification(userID, fromUserID, message string) error {
	model := models.NewConn()
	idObj, err := primitive.ObjectIDFromHex(fromUserID)
	if err != nil {
		return err
	}

	fromUser, err := model.GetUserByID(idObj)
	if err != nil {
		log.Println("Error fetching user:", err)
		return err
	}

	notification := socket.Notification{
		FromUserID: fromUserID,
		ToUserID:   userID,
		Message:    message,
		Created_At: time.Now(),
	}

	// Cập nhật thông báo với thông tin chi tiết về người gửi
	notificationDetail := gin.H{
		"from_user_id": fromUser.ID,
		"from_name":    fromUser.Name,
		"from_avatar":  fromUser.Avatar,
		"message":      notification.Message,
		"created_at":   notification.Created_At,
	}

	filter := bson.M{
		"from_user_id": fromUserID,
		"to_user_id":   userID,
		"message":      message,
		"avatar":       fromUser.Avatar,
		"read":         false,
		"created_at":   time.Now(),
	}
	_, err = model.CollectionNotify.InsertOne(context.Background(), filter)
	if err != nil {
		return err
	}

	notifyMutex.Lock()
	defer notifyMutex.Unlock()

	if conn, exists := notifyClients[userID]; exists {
		err := conn.WriteJSON(notificationDetail)
		if err != nil {
			fmt.Println("Error sending notification:", err)
			conn.Close()
			delete(notifyClients, userID)
		}
		fmt.Println("Gửi thành công đến " + userID)
	} else {
		fmt.Println("No WebSocket connection found for user:", userID)
	}
	return nil
}

type ReqDataComment struct {
	IDBlog  string `json:"blog_id" bson:"blog_id"`
	Message string `json:"message" bson:"message"`
	UserID  string `json:"user_id" bson:"user_id"`
	Avatar  string `json:"avatar" bson:"avatar"`
	Name    string `json:"name" bson:"name"`
}

func SOcketCOmmentBlog(c *gin.Context) {
	var data ReqDataComment
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err":    err.Error(),
			"status": http.StatusBadRequest,
			"msg":    "Failed to bind request body",
		})
	}
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":    "Failed to connect websocket",
			"status": http.StatusInternalServerError,
			"err":    err.Error(),
		})
		return
	}
	defer conn.Close()

	blogID := data.IDBlog
	blogCommentClientsMutex.Lock()
	if _, ok := blogCommentClients[blogID]; !ok {
		blogCommentClients[blogID] = &BlogCommentClients{
			Clients: make(map[string]*websocket.Conn),
			Mutex:   sync.Mutex{},
		}
	}
	blogCommentClientsMutex.Unlock()

	blogCommentClients[blogID].Mutex.Lock()
	blogCommentClients[blogID].Clients[c.Request.RemoteAddr] = conn
	blogCommentClients[blogID].Mutex.Unlock()
	dbConn := models.NewConn()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			blogCommentClients[blogID].Mutex.Lock()
			delete(blogCommentClients[blogID].Clients, c.Request.RemoteAddr)
			blogCommentClients[blogID].Mutex.Unlock()
			break
		}
		if string(message) == "get_comment" {
			idObj, err := primitive.ObjectIDFromHex(data.IDBlog)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Invalid user ID"}`))
				continue
			}

			listCOmment, err := dbConn.GetAllCommentByBlog(idObj)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Failed to fetch notifications"}`))
				continue
			}

			notificationsJSON, _ := json.Marshal(listCOmment)
			conn.WriteMessage(websocket.TextMessage, notificationsJSON)
		}
		if string(message) == "new_comment" {
			idUserObj, _ := primitive.ObjectIDFromHex(data.UserID)

			idObj, err := primitive.ObjectIDFromHex(data.IDBlog)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Invalid user ID"}`))
				continue
			}
			newComment, err := dbConn.CommentByBlog(idObj, idUserObj, data.Message)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Failed to fetch comment"}`))
				continue
			}
			newCommentJSON, err := json.Marshal(newComment)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Failed to marshal comment"}`))
				continue
			}

			blogCommentClientsMutex.Lock()
			blogCommentClients[blogID].Mutex.Lock()
			for _, clientConn := range blogCommentClients[blogID].Clients {
				clientConn.WriteMessage(websocket.TextMessage, newCommentJSON)
			}
			blogCommentClients[blogID].Mutex.Unlock()
		}
	}
}
