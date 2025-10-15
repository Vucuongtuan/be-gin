package models

import (
	"be/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Login struct {
	ID           *primitive.ObjectID `json:"_id" bson:"_id"`
	Name_Account string              `json:"name_account" bson:"name_account"`
	Password     string              `json:"password" bson:"password"`
}

type Online struct {
	ID        *primitive.ObjectID `json:"_id" bson:"_id"`
	Token     string              `json:"token" bson:"token"`
	UserID    string              `json:"user_id" bson:"user_id"`
	Name      string              `json:"name" bson:"name"`
	Email     string              `json:"email" bson:"email"`
	Online_at *time.Time          `json:"online_at" bson:"online_at"`
}
type Otp struct {
	ID         *primitive.ObjectID `json:"_id " bson:"_id"`
	Email      string              `json:"email" bson:"email"`
	Otp        int64               `json:"otp" bson:"otp"`
	Created_At time.Time          `json:"created_at" bson:"created_at"`
}

func CheckAccountLogin(data Login, conn Conn, c *gin.Context) (Login, error) {
	dbAccount := Login{}
	filter := bson.M{
		"name_account": data.Name_Account,
	}
	err := conn.CollectionAccount.FindOne(context.Background(), filter).Decode(&dbAccount)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Login{}, fmt.Errorf("Account not found: %v", err)
		}
		return Login{}, err
	}

	err = utils.CheckPasswordHash(data.Password, dbAccount.Password)
	if err != nil {
		return Login{}, err
	}
	return dbAccount, nil
}
func (conn *Conn) AddOtp(otp int64, email string) error {
    _,err := conn.CollectionOtp.InsertOne(context.Background(), bson.M{
        "otp":        otp,
        "email":      email,
        "created_at": time.Now(),
    })
    if err != nil {
        return err
    }
    return nil
}

func  ResetPassword(conn *Conn,otp int64, email string, password string) (int64, string, error) {
	
	var OtpStore Otp
	filterCheckOtp := bson.M{
		"email": email,
	}
	findOptions := options.FindOne().SetSort(bson.D{{"created_at", -1}})
	err := conn.CollectionOtp.FindOne(context.Background(), filterCheckOtp,findOptions).Decode(&OtpStore)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return http.StatusNotFound, "otp not found", errors.New("OTP not found")
		}
		return http.StatusBadRequest, "Can't check otp , please try again", err
	}

	
	if OtpStore.Otp != otp {
		fmt.Println(OtpStore.Otp , otp)
		return http.StatusUnauthorized, "Invalid OTP", errors.New("Invalid OTP")
	}

	var user User
    err = conn.CollectionUser.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return http.StatusNotFound,"Can't find user",err
		}
		return http.StatusInternalServerError,"Can't query user from database ,please try again",err
	}

	hash, _ := utils.HashPassword(password)
	objectId, err := primitive.ObjectIDFromHex(user.Account)
    if err != nil {
         	return http.StatusInternalServerError, "Can't update password, please try again", err
    }
	update := bson.M{
		"$set": bson.M{
			"password": hash,
			"updated_at":       time.Now(),
		},
	}
	_, err = conn.CollectionAccount.UpdateOne(context.Background(), bson.M{"_id":objectId}, update)
	if err != nil {
		return http.StatusInternalServerError, "Can't update password, please try again", err
	}
	return http.StatusOK, "Password updated successfully", nil
}

func  LogoutAccount(token string )error {
	conn := NewConn()
	_,err := conn.CollectionOnline.DeleteOne(context.Background(),bson.M{"token": token})
	if err != nil {
		return err
	}
	return nil
}