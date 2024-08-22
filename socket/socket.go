package socket

import "time"

var Broadcast = make(chan Notification)

type Notification struct {
	FromUserID string    `json:"from_user_id" bson:"from_user_id"`
	ToUserID   string    `json:"to_user_id" bson:"to_user_id"`
	Message    string    `json:"message" bson:"message"`
	Created_At time.Time `json:"created_at" bson:"created_at"`
}
