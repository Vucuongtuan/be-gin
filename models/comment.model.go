package models

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID         *primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	Avatar     string              `bson:"avatar" json:"avatar,omitempty"`
	UserName   string              `bson:"username" json:"username"`
	BlogID     primitive.ObjectID  `bson:"blog_id" json:"blog_id"`
	Content    string              `bson:"content" json:"content"`
	UserID     primitive.ObjectID  `bson:"user_id" json:"user_id"`
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

func (conn *Conn) GetCommentByBlog(blogID primitive.ObjectID) ([]Comment, error) {
	var comments []Comment
	cur, err := conn.CollectionComments.Find(context.Background(), bson.M{"blog_id": blogID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var comment Comment
		if err := cur.Decode(&comment); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (conn *Conn) CommentByBlog(blogID primitive.ObjectID, userID primitive.ObjectID, message string) (*Comment, error) {
	var dataUser User
	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&dataUser)
	if err != nil {
		return nil, err
	}
	var user User
	err = conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	username := user.Name
	now := time.Now()

	comment := &Comment{
		BlogID:     blogID,
		UserName:   username,
		Avatar:     dataUser.Avatar,
		Content:    message,
		UserID:     userID,
		Created_At: &now,
		Updated_At: &now,
	}
	filter := bson.M{
		"blog_id":    blogID,
		"user_id":    userID,
		"username":   username,
		"avatar":     dataUser.Avatar,
		"content":    message,
		"created_at": &now,
		"updated_at": &now,
	}
	_, err = conn.CollectionComments.InsertOne(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	return comment, nil
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
func (conn *Conn) GetAllCommentByBlog(idBlog primitive.ObjectID) ([]Comment, error) {
	filter := bson.M{
		"blog_id": idBlog,
	}
	var comments []Comment
	cursor, err := conn.CollectionComments.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var comment Comment
		if err := cursor.Decode(&comment); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if comments == nil {
		return nil, errors.New("comment not found")
	}
	return comments, nil
}
