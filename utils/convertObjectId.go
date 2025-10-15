package utils

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ConvertStringToObjectID(idStr string) (primitive.ObjectID, error) {
    objectID, err := primitive.ObjectIDFromHex(idStr)
    if err != nil {
        return primitive.NilObjectID, err
    }
    return objectID, nil
}
