package helpers

import (
	"fmt"
	"github.com/gorilla/websocket"
)

func NotifyFollower(conn *websocket.Conn, name string) error {
	message := fmt.Sprintf("%s has started following you.", name)
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return err
	}
	return nil
}
