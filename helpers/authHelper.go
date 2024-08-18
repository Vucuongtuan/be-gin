package helpers

import (
	"be/models"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func AuthUser(dbAccount models.Login)(models.User,error) {
	var TUser models.User
	var conn models.Conn
	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id":dbAccount.ID}).Decode(&TUser)
	if err != nil {
		return models.User{},err
	}
	return TUser,nil
}

func CheckEmail(email string,conn *models.Conn)(bool,error){
	err := conn.CollectionUser.FindOne(context.Background(),bson.M{"email":email})
	if err != nil {
		return false, fmt.Errorf("Email already exists ")
	}
	return true, nil
}
func CheckAccountName(nameAccount string, conn *models.Conn) (bool, error) {
	 err := conn.CollectionAccount.FindOne(context.Background(),bson.M{
		"name_account":nameAccount,
	 })
	if err != nil {
		return false, fmt.Errorf("name_account already exists ")
	}
	return true, nil
}