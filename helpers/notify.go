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
var commentClient = make(map[string][]*websocket.Conn)
var commentMutex = sync.Mutex{}

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

func GetCommmentByBlog(c *gin.Context) {
	idBlog := c.Param("blogID")
	if idBlog == "" {
		fmt.Println("Missing blogID")
		return
	}
	fmt.Println("blog : ", idBlog)
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":    "Failed to upgrade connection",
			"status": http.StatusInternalServerError,
			"err":    err.Error(),
		})
		return
	}

	defer conn.Close()
	commentMutex.Lock()
	if commentClient[idBlog] == nil {
		commentClient[idBlog] = []*websocket.Conn{}
	}
	commentClient[idBlog] = append(commentClient[idBlog], conn)
	commentMutex.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			commentMutex.Lock()
			commentClient[idBlog] = RemoveConnection(commentClient[idBlog], conn)
			if len(commentClient[idBlog]) == 0 {
				delete(notifyClients, idBlog)
			}
			commentMutex.Unlock()
			break
		}

	}

}
func SendCommentAllUser(idBlog string, userID string, name, avatar, message string) error {
	comment := gin.H{
		"blog_id":    idBlog,
		"user_id":    userID,
		"username":   name,
		"content":    message,
		"avatar":     avatar,
		"created_at": time.Now(),
	}

	commentMutex.Lock()
	defer commentMutex.Unlock()

	if clients, exists := commentClient[idBlog]; exists {
		for _, conn := range clients {
			err := conn.WriteJSON(comment)
			if err != nil {
				fmt.Println("Error sending notification:", err)
				conn.Close()
				commentClient[idBlog] = RemoveConnection(commentClient[idBlog], conn)
				if len(commentClient[idBlog]) == 0 {
					delete(commentClient, idBlog)
				}
			}
		}
	} else {
		fmt.Println("No WebSocket connection found for blog:", idBlog)
	}
	return nil
}
func RemoveConnection(conns []*websocket.Conn, connToRemove *websocket.Conn) []*websocket.Conn {
	var updatedConns []*websocket.Conn
	for _, conn := range conns {
		if conn != connToRemove {
			updatedConns = append(updatedConns, conn)
		}
	}
	return updatedConns
}
