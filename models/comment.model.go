package models

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID         *primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	UserName   *string             `bson:"username" json:"username"`
	BlogID     *primitive.ObjectID `bson:"blog_id" json:"blog_id"`
	Content    string              `bson:"content" json:"content"`
	UserID     *primitive.ObjectID `bson:"user_id" json:"user_id"`
	Like       []Like              `bson:"like" json:"like"`
	DisLike    []Dislike           `bson:"dislike" json:"dislike"`
	Created_At *time.Time          `bson:"created_at" json:"created_at"`
	Updated_At *time.Time          `bson:"updated_at" json:"updated_at"`
}
type Reply struct {
	ID         *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserName   *string             `bson:"username" json:"username"`
	CommentID  *primitive.ObjectID `bson:"comment_id" json:"comment_id"`
	Content    string              `bson:"content" json:"content"`
	UserID     *primitive.ObjectID `bson:"userId" json:"user_id"`
	Like       []Like              `bson:"like" json:"like"`
	DisLike    []Dislike           `bson:"dislike" json:"dislike"`
	Created_At *time.Time          `bson:"created_at" json:"created_at"`
	Updated_At *time.Time          `bson:"updated_at" json:"updated_at"`
}

func (conn *Conn) CommentByBlog(blogID primitive.ObjectID, userID primitive.ObjectID, message string) (int, string, error) {
	var user User

	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return http.StatusBadRequest, "Can't find user", err
	}
	username := user.Name
	now := time.Now()
	filter := bson.M{
		"blog_id":    blogID,
		"username":   username,
		"content":    message,
		"user_id":    userID,
		"created_at": now,
		"updated_at": now,
	}
	_, err = conn.CollectionComments.InsertOne(context.Background(), filter)
	if err != nil {
		return http.StatusInternalServerError, "Can't insert comment , please try again later", err
	}
	return http.StatusOK, "Comment inserted successfully", nil
}

func (conn *Conn) ReplyComment(blogID primitive.ObjectID, userID primitive.ObjectID, commentID primitive.ObjectID, message string) (int64, string, error) {
	var user User

	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return http.StatusBadRequest, "Can't find user", err
	}
	username := user.Name
	filter := bson.M{
		"comment_id": commentID,
		"username":   username,
		"comment":    message,
		"user_id":    userID,
	}
	_, err = conn.CollectionReplyComments.InsertOne(context.Background(), filter)
	if err != nil {
		return http.StatusInternalServerError, "Can't insert reply comment , please try again later", err
	}
	return http.StatusOK, "Reply comment inserted successfully", nil
}

func (conn *Conn) CommentLike(userID primitive.ObjectID, CommentID primitive.ObjectID) error {
	var user User

	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}
	username := user.Name
	filter := bson.M{
		"$push": bson.M{
			"like":       userID,
			"username":   username,
			"created_at": time.Now(),
		},
	}
	_, err = conn.CollectionComments.UpdateByID(context.Background(), CommentID, filter)
	if err != nil {
		return err
	}
	return nil
}
func (conn *Conn) CommentDisLike(userID primitive.ObjectID, CommentID primitive.ObjectID) error {
	var user User

	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}
	username := user.Name
	filter := bson.M{
		"$push": bson.M{
			"dislike":    userID,
			"username":   username,
			"created_at": time.Now(),
		},
	}
	_, err = conn.CollectionComments.UpdateByID(context.Background(), CommentID, filter)
	if err != nil {
		return err
	}
	return nil
}
func (conn *Conn) ReplyCommentLike(userID primitive.ObjectID, CommentID primitive.ObjectID) error {
	var user User

	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}
	username := user.Name
	filter := bson.M{
		"$push": bson.M{
			"like":       userID,
			"username":   username,
			"created_at": time.Now(),
		},
	}
	filterByCommentID := bson.M{
		"comment_id": CommentID,
	}
	_, err = conn.CollectionComments.UpdateOne(context.Background(), filterByCommentID, filter)
	if err != nil {
		return err
	}
	return nil
}
func (conn *Conn) ReplyCommentDisLike(userID primitive.ObjectID, CommentID primitive.ObjectID) error {
	var user User

	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}
	username := user.Name
	filter := bson.M{
		"$push": bson.M{
			"dislike":    userID,
			"username":   username,
			"created_at": time.Now(),
		},
	}
	filterByCommentID := bson.M{
		"comment_id": CommentID,
	}
	_, err = conn.CollectionComments.UpdateOne(context.Background(), filterByCommentID, filter)
	if err != nil {
		return err
	}
	return nil
}
