package models

import (
	"be/config"

	"go.mongodb.org/mongo-driver/mongo"
)

type Conn struct {
	CollectionUser          *mongo.Collection
	CollectionAccount       *mongo.Collection
	CollectionBlogs         *mongo.Collection
	CollectionOnline        *mongo.Collection
	CollectionHashtags      *mongo.Collection
	CollectionOtp           *mongo.Collection
	CollectionComments      *mongo.Collection
	CollectionReplyComments *mongo.Collection
	CollectionNotify        *mongo.Collection
}

func NewConn() *Conn {
	return &Conn{
		CollectionUser:          config.GetCollection("users"),
		CollectionAccount:       config.GetCollection("accounts"),
		CollectionBlogs:         config.GetCollection("blogs"),
		CollectionOnline:        config.GetCollection("online"),
		CollectionHashtags:      config.GetCollection("hashtags"),
		CollectionOtp:           config.GetCollection("otp"),
		CollectionComments:      config.GetCollection("comments"),
		CollectionReplyComments: config.GetCollection("reply_comments"),
		CollectionNotify:        config.GetCollection("notify"),
	}
}
