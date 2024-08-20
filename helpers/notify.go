package helpers

import (
	"be/models"
	"be/socket"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"sync"
	"time"
)

var notifyClients = make(map[string]*websocket.Conn)
var notifyMutex = sync.Mutex{}

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

	notifyMutex.Lock()
	if conn, exists := notifyClients[userID]; exists {
		err := conn.WriteJSON(notificationDetail)
		if err != nil {
			log.Println("Error sending notification:", err)
			conn.Close()
			delete(notifyClients, userID)
		}
	}
	notifyMutex.Unlock()
	return nil
}
