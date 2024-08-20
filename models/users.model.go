package models

import (
	"be/socket"
	"be/utils"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID            *primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name          string              `json:"name" bson:"name"`
	Email         string              `json:"email" bson:"email"`
	Date_BirthDay *time.Time          `json:"date_birth" bson:"date_birth,omitempty"`
	Avatar        *string             `json:"avatar" bson:"avatar,omitempty"`
	Account       string              `json:"account" bson:"account"`
	Followers     *[]string           `json:"followers" bson:"followers,omitempty"`
	Follow        *[]string           `json:"follow" bson:"follow,omitempty"`
	Created_At    *time.Time          `json:"created_at" bson:"created_at,omitempty"`
	Updated_At    *time.Time          `json:"updated_at" bson:"updated_at,omitempty"`
}

type Account struct {
	Name_Account string     `json:"name_account" bson:"name_account"`
	Password     string     `json:"password_account" bson:"password"`
	Created_At   *time.Time `json:"create_at" bson:"create_at"`
	Updated_At   *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
}
type CreateUser struct {
	Name        string     `json:"name" bson:"name"`
	Email       string     `json:"email" bson:"email"`
	DateBirth   *time.Time `json:"date_birth" bson:"date_birth"`
	NameAccount string     `json:"name_account" bson:"name_account"`
	Password    string     `json:"password" bson:"password"`
	CreatedAt   *time.Time `json:"created_at" bson:"created_at"`
}

type UPdateUser struct {
	Name          string     `json:"name" bson:"name"`
	Email         string     `json:"email" bson:"email"`
	Date_BirthDay *time.Time `json:"date_birth" bson:"date_birth"`
	Avatar        *string    `json:"avatar" bson:"avatar"`
	Update_At     *time.Time `json:"update_at" bson:"update_at"`
}

func (c *Conn) GetAllUsers(page int64, limit int64) ([]User, int64, int64, error) {
	var users []User

	skip := (page - 1) * limit
	option := options.Find()
	option.SetSkip(skip)
	option.SetLimit(limit)
	get, err := c.CollectionUser.Find(context.Background(), bson.M{}, option)
	if err != nil {
		return nil, 0, 0, err
	}

	defer get.Close(context.Background())
	for get.Next(context.Background()) {
		var user User
		if err = get.Decode(&user); err != nil {
			return nil, 0, 0, err
		}
		users = append(users, user)
	}
	if err = get.Err(); err != nil {
		return nil, 0, 0, err
	}
	total, err := c.CollectionUser.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return nil, 0, 0, err
	}
	totalPages := (total + limit - 1) / limit
	return users, total, totalPages, nil
}

func (c *Conn) CreateUser(user User) error {

	filter := bson.M{
		"name":       user.Name,
		"email":      user.Email,
		"account":    user.Account,
		"follow":     []string{},
		"followers":  []string{},
		"save_blog":  []string{},
		"created_at": time.Now().UTC(),
		"updated_at": time.Now().UTC(),
	}
	_, err := c.CollectionUser.InsertOne(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (c *Conn) CreateAccount(account Account) (primitive.ObjectID, error) {

	hash, _ := utils.HashPassword(account.Password)
	filter := bson.M{
		"name_account": account.Name_Account,
		"password":     hash,
		"created_at":   time.Now().UTC(),
		"updated_at":   time.Now().UTC(),
	}
	create, err := c.CollectionAccount.InsertOne(context.Background(), filter)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return create.InsertedID.(primitive.ObjectID), nil
}
func (c *Conn) DeleteUser(_id primitive.ObjectID) error {

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": _id}
	_, err := c.CollectionAccount.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	stringId := _id.Hex()
	filterUser := bson.M{"account": stringId}
	_, err = c.CollectionUser.DeleteOne(context.Background(), filterUser)
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) UpdateUser(_id primitive.ObjectID, data UPdateUser) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filterId := bson.M{"_id": _id}
	filterUpdate := bson.M{
		"$set": bson.M{
			"name":       data.Name,
			"email":      data.Email,
			"date_birth": data.Date_BirthDay,
			"avatar":     data.Avatar,
			"update_at":  data.Update_At,
		},
	}
	_, err := c.CollectionUser.UpdateOne(context.Background(), filterId, filterUpdate)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Conn) UnFollow(id string, idFollowers string) error {
	filter := bson.M{
		"$pull": bson.M{
			"followers": bson.M{
				"idFollower": idFollowers,
			},
		},
	}
	_, err := conn.CollectionUser.UpdateOne(context.Background(), bson.M{"_id": id}, filter)
	if err != nil {
		return err
	}

	filterFollower := bson.M{
		"$pull": bson.M{
			"follow": bson.M{
				"idFollow": id,
			},
		},
	}
	_, err = conn.CollectionUser.UpdateOne(context.Background(), bson.M{"_id": idFollowers}, filterFollower)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Conn) Follow(id string, idFollow string, name string) error {
	now := time.Now()
	filter := bson.M{
		"$push": bson.M{
			"followers": bson.M{
				"name":       name,
				"idFollower": idFollow,
				"created_at": &now,
			},
		},
	}
	_, err := conn.CollectionUser.UpdateOne(context.Background(), bson.M{"_id": id}, filter)
	if err != nil {
		return err
	}
	filterFollower := bson.M{
		"$push": bson.M{
			"follow": bson.M{
				"name":       name,
				"idFollow":   idFollow,
				"created_at": &now,
			},
		},
	}
	_, err = conn.CollectionUser.UpdateOne(context.Background(), bson.M{"_id": idFollow}, filterFollower)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Conn) SaveBlogs(id string, idBlog string) error {
	filter := bson.M{
		"$set": bson.M{
			"save_blog": idBlog,
		},
	}

	_, err := conn.CollectionUser.UpdateOne(context.Background(), bson.M{"_id": id}, filter)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Conn) GetUserByID(id primitive.ObjectID) (User, error) {
	var user User

	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (conn *Conn) NotifyModel(idUser primitive.ObjectID, idAuthor primitive.ObjectID, name string, title *string, description *string) error {

	filter := bson.M{
		"name":        name,
		"id_user":     idUser,
		"title":       title,
		"description": description,
		"created_at":  time.Now(),
	}
	_, err := conn.CollectionNotify.InsertOne(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Conn) GetRecentNotificationsByUserID(userID primitive.ObjectID) ([]socket.Notification, error) {
	filter := bson.M{"to_user_id": userID}
	options := options.Find().SetSort(bson.D{{"create_at", -1}}).SetLimit(10) // Lấy 10 thông báo mới nhất

	cursor, err := conn.CollectionNotify.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var notifications []socket.Notification
	if err := cursor.All(context.Background(), &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

// GetAllNotificationsByUserID lấy tất cả thông báo cho user từ một vị trí bắt đầu
func (conn *Conn) GetAllNotificationsByUserID(userID primitive.ObjectID) ([]socket.Notification, error) {
	filter := bson.M{
		"to_user_id": userID,
	}
	options := options.Find().SetSort(bson.D{{"create_at", -1}}) // Sắp xếp theo thời gian tạo

	cursor, err := conn.CollectionNotify.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var notifications []socket.Notification
	if err := cursor.All(context.Background(), &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}
