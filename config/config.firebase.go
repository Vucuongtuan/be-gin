package config

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)


func ConfigFirebaseStorage()(*storage.Client,error){
	optURL := option.WithCredentialsFile("./bechatapptc-firebase-adminsdk-e3tfr-28fb9a8628.json")

	storageConn,err :=storage.NewClient(context.Background(),optURL)
	if err != nil{
		return nil,err
	}
	return storageConn,nil
}
