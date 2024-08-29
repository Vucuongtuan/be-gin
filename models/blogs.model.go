package models

import (
	"be/utils"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"
)

type Blogs struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	Author      primitive.ObjectID `bson:"author" json:"author"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Slug        *string            `bson:"slug" json:"slug"`
	File        *string            `bson:"file" json:"file"`
	Like        *[]Like            `bson:"like" json:"like"`
	Hashtags    []Hashtags         `bson:"hashtags" json:"hashtags"`
	Link        *string            `bson:"link" json:"link"`
	Type        *string            `bson:"type" json:"type"`
	DisLike     *[]Dislike         `bson:"dislike" json:"dislike"`
	Share       *[]Share           `bson:"share" json:"share"`
	View        *int64             `bson:"view" json:"view"`
	Created_At  *time.Time         `bson:"created_at" json:"created_at"`
	Updated_At  *time.Time         `bson:"updated_at" json:"updated_at"`
}
type Dislike struct {
	ID         *primitive.ObjectID `bson:"_id" json:"_id"`
	Name       string              `bson:"name" json:"name"`
	UserID     *primitive.ObjectID `bson:"user_id" json:"user_id"`
	Created_At *time.Time          `bson:"created_at" json:"created_at"`
}
type Share struct {
	ID         *primitive.ObjectID `bson:"_id" json:"_id"`
	Name       string              `bson:"name" json:"name"`
	UserID     *primitive.ObjectID `bson:"user_id" json:"user_id"`
	Created_At *time.Time          `bson:"created_at" json:"created_at"`
}
type Like struct {
	ID         *primitive.ObjectID `bson:"_id" json:"_id"`
	Name       string              `bson:"name" json:"name"`
	UserID     *primitive.ObjectID `bson:"user_id" json:"user_id"`
	Created_At *time.Time          `bson:"created_at" json:"created_at"`
}
type Hashtags struct {
	ID         *primitive.ObjectID `bson:"_id" json:"_id"`
	Name       string              `bson:"name" json:"name"`
	Slug       string              `bson:"slug" json:"slug"`
	Count      int64               `bson:"count" json:"count"`
	Created_At *time.Time          `bson:"created_at" json:"created_at"`
	Updated_At *time.Time          `bson:"updated_at" json:"updated_at"`
}
type CreateBlogsDto struct {
	Title       string   `bson:"title" json:"title" form:"title"`
	Description string   `bson:"description" json:"description" form:"description"`
	Hashtags    []string `bson:"hashtags" json:"hashtags" form:"hashtags"`
	Link        *string  `bson:"link" json:"link" form:"link"`
	Type        *string  `bson:"type" json:"type" form:"type"`
}

type Comments struct {
	ID         *primitive.ObjectID `bson:"_id" json:"_id"`
	Message    string              `bson:"message" json:"message"`
	Like       int                 `bson:"like" json:"like"`
	Dislike    int                 `bson:"dislike" json:"dislike"`
	Blog       primitive.ObjectID  `bson:"blog" json:"blog"`
	Created_At time.Time           `bson:"created_at" json:"created_at"`
	Updated_At time.Time           `bson:"updated_at" json:"updated_at"`
}

// params page(int) and limit(int)
// des : get all blog
func (conn *Conn) GetAllBlogs(page int64, limit int64) ([]Blogs, int64, int64, error) {
	var blogs []Blogs
	skip := (page - 1) * limit
	option := options.Find()
	option.SetSort(bson.M{"created_at": -1})
	option.SetSkip(skip)
	option.SetLimit(limit)
	get, err := conn.CollectionBlogs.Find(context.Background(), bson.M{}, option)
	if err != nil {

		return nil, 0, 0, err
	}

	defer get.Close(context.Background())
	for get.Next(context.Background()) {
		var blog Blogs
		if err = get.Decode(&blog); err != nil {
			return nil, 0, 0, err
		}
		blogs = append(blogs, blog)
	}
	if err = get.Err(); err != nil {
		return nil, 0, 0, err
	}
	total, err := conn.CollectionBlogs.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return nil, 0, 0, err
	}
	totalPages := (total + limit - 1) / limit
	return blogs, total, totalPages, nil

}

// params : page(int) and limit(int)
// des : func get blog new featured images from the past 2 months
func (conn *Conn) GetBlogNewFeatured(page int64, limit int64) ([]Blogs, int64, int64, error) {

	realTime := time.Now()                     // real time
	twoMonthsAgo := realTime.AddDate(0, -2, 0) // get two months ago

	// filter query blog two months ago
	filterQuery := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{
					"$gte": twoMonthsAgo,
				},
			},
		},
		{
			"$sort": bson.M{
				"viewCount": -1,
			},
		},
		{
			"$skip": (page - 1) * limit,
		},
		{
			"$limit": limit,
		},
	}
	cursor, err := conn.CollectionBlogs.Aggregate(context.Background(), filterQuery)
	if err != nil {
		return nil, 0, 0, err
	}
	defer cursor.Close(context.Background())

	var blogs []Blogs
	err = cursor.All(context.Background(), &blogs)
	if err != nil {
		return nil, 0, 0, err
	}

	totalCount, err := conn.CollectionBlogs.CountDocuments(context.Background(), bson.M{
		"createdAt": bson.M{
			"$gte": twoMonthsAgo,
		},
	})
	if err != nil {
		return nil, 0, 0, err
	}

	return blogs, totalCount, limit, nil
}

func (conn *Conn) CreateBlog(createBlogDto CreateBlogsDto, userIDOnject primitive.ObjectID, c *gin.Context) (int, string, error) {
	timeNow := time.Now().UTC()
	fileLinks, exists := c.Get("file_link")
	if !exists {
		return http.StatusInternalServerError, "File links not found", nil
	}
	var generateSlug string
	if createBlogDto.Title == "" {
		nameTitle := timeNow.Format("20060102150405") + userIDOnject.Hex()
		generateSlug = utils.GenerateSlug(nameTitle, 20)
	} else {
		generateSlug = utils.GenerateSlug(createBlogDto.Title, 20)
	}
	var hashtags []Hashtags
	for _, hashtag := range createBlogDto.Hashtags {
		now := time.Now().UTC()
		genarateSlugg := utils.GenerateSlug(hashtag, 20)

		// Tìm hashtag trong cơ sở dữ liệu
		filter := bson.M{"name": hashtag}
		update := bson.M{
			"$set": bson.M{
				"updated_at": &now,
			},
			"$inc": bson.M{"count": 1},
		}
		options := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

		var existingHashtag Hashtags
		err := conn.CollectionHashtags.FindOneAndUpdate(context.Background(), filter, update, options).Decode(&existingHashtag)
		if err != nil && err != mongo.ErrNoDocuments {
			return http.StatusBadRequest, "Error updating hashtag", err
		}
		if err == mongo.ErrNoDocuments {
			hashtagDoc := Hashtags{
				Name:       hashtag,
				Slug:       genarateSlugg,
				Created_At: &now,
				Updated_At: &now,
				Count:      1,
			}
			fmt.Println(hashtagDoc)
			_, err := conn.CollectionHashtags.InsertOne(context.Background(), hashtagDoc)
			if err != nil {
				return http.StatusBadRequest, "Error inserting hashtag", err
			}
			hashtags = append(hashtags, hashtagDoc)
		} else {
			hashtags = append(hashtags, existingHashtag)
		}
	}
	now := time.Now().UTC()
	// authorObjectID ,_ := primitive.ObjectIDFromHex(userID)
	blog := bson.M{
		"author":      userIDOnject,
		"title":       createBlogDto.Title,
		"description": createBlogDto.Description,
		"slug":        generateSlug + "-" + now.Format("20060102150405"),
		"file":        fileLinks,
		"hashtags":    hashtags,
		"like":        bson.A{},
		"dislike":     bson.A{},
		"share":       bson.A{},
		"link":        createBlogDto.Link,
		"type":        createBlogDto.Type,
		"view":        new(int64),
		"created_at":  &now,
		"updated_at":  &now,
	}

	_, err := conn.CollectionBlogs.InsertOne(context.Background(), blog)
	if err != nil {
		fmt.Println("err : " + err.Error())
		return http.StatusBadRequest, "Error inserting blog", err
	}

	return http.StatusCreated, "Created Blog successfully", nil
}

func (conn *Conn) UpdateBlog(id string, updateBlogDto Blogs) (int, string, error) {
	update := bson.M{}
	v := reflect.ValueOf(updateBlogDto)
	t := reflect.TypeOf(updateBlogDto)
	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		// fieldValue := v.Field(i).Interface()

		switch fieldName {
		case "Title":
			if updateBlogDto.Title != "" {
				update["title"] = updateBlogDto.Title
			}
		case "Description":
			if updateBlogDto.Description != "" {
				update["description"] = updateBlogDto.Description
			}
		case "Slug":
			if updateBlogDto.Slug != nil && *updateBlogDto.Slug != "" {
				update["slug"] = *updateBlogDto.Slug
			}
		case "Hashtags":
			if len(updateBlogDto.Hashtags) > 0 {
				update["hashtags"] = updateBlogDto.Hashtags
			}
		case "File":
			if updateBlogDto.File != nil && *updateBlogDto.File != "" {
				update["file"] = updateBlogDto.File
			}

		case "Update_At":
			if updateBlogDto.Updated_At != nil {
				update["update_at"] = time.Now().UTC()
			}
		}
	}
	_, err := conn.CollectionBlogs.UpdateByID(context.Background(), id, bson.M{"$set": update})
	if err != nil {
		return http.StatusBadRequest, "Error updating blog ", err
	}
	return http.StatusOK, "Update Blog successfully", err
}
func (conn *Conn) DeleteBlog(id string) (int, string, error) {
	result, err := conn.CollectionBlogs.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return http.StatusInternalServerError, "Error deleting blog ", err
	}
	if result.DeletedCount == 0 {
		return http.StatusNotFound, "Can't delete blog because blog not found", nil
	}

	return http.StatusOK, "Delete blog successfully", nil
}

func (conn *Conn) GetBlogDetailBySlug(slug string) (int, string, any, error) {
	var blogs Blogs
	err := conn.CollectionBlogs.FindOne(context.Background(), bson.M{"slug": slug}).Decode(&blogs)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return http.StatusNotFound, "Can't get blog details", Blogs{}, err
		}
		return http.StatusInternalServerError, "Can not find blog details", Blogs{}, err
	}
	fmt.Println("author :", blogs.Author)
	var user User
	err = conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": blogs.Author}).Decode(&user)
	if err != nil {
		return http.StatusInternalServerError, "Can not find user details", Blogs{}, err
	}

	data := struct {
		Blog *Blogs `json:"blog"`
		User *User  `json:"author"`
	}{
		Blog: &blogs,
		User: &user,
	}
	return http.StatusOK, "Get blog details successfully", data, nil
}
func (conn *Conn) LikeBlog(userID primitive.ObjectID, blogID primitive.ObjectID) error {
	var user User

	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}
	username := user.Name
	now := time.Now()
	filter := bson.M{
		"$push": bson.M{
			"like": bson.M{
				"user_id":    userID,
				"name":       username,
				"created_at": now,
			},
		},
	}
	_, err = conn.CollectionBlogs.UpdateByID(context.Background(), blogID, filter)
	if err != nil {
		return err
	}
	return nil
}
func (conn *Conn) DisLikeBlog(userID primitive.ObjectID, blogID primitive.ObjectID) error {
	var user User

	err := conn.CollectionUser.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}
	username := user.Name
	filter := bson.M{
		"$push": bson.M{
			"dislike":    userID,
			"name":       username,
			"created_at": time.Now(),
		},
	}
	_, err = conn.CollectionBlogs.UpdateByID(context.Background(), blogID, filter)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Conn) View(slug string) error {
	filter := bson.M{
		"$inc": bson.M{
			"view": 1,
		},
	}
	_, err := conn.CollectionBlogs.UpdateOne(context.Background(), bson.M{"slug": slug}, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("update failed: no documents found with the specified blog ID")
		}
		return err
	}
	return nil
}

func (conn *Conn) GetBlogByAuthor(id primitive.ObjectID, page int) ([]Blogs, int, string, error) {
	var blogs []Blogs
	limitStr := os.Getenv("LIMIT")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	skip := (page - 1) * limit

	filter := bson.M{
		"author": id,
	}

	options := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))

	blogsData, err := conn.CollectionBlogs.Find(context.Background(), filter, options)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, http.StatusNotFound, "blog not found", nil
		}
		return nil, http.StatusInternalServerError, "Cannot find blog details", err
	}
	defer blogsData.Close(context.Background())

	for blogsData.Next(context.Background()) {
		var blog Blogs
		if err := blogsData.Decode(&blog); err != nil {
			return nil, http.StatusInternalServerError, "Error decoding blog data", err
		}
		blogs = append(blogs, blog)
	}

	if err := blogsData.Err(); err != nil {
		return nil, http.StatusInternalServerError, "Error iterating over blog cursor", err
	}

	totalBlogs, err := conn.CollectionBlogs.CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, http.StatusInternalServerError, "Error getting total blog count", err
	}

	return blogs, http.StatusOK, fmt.Sprintf("Total blogs: %d", totalBlogs), nil
}
func (conn *Conn) SearchBlogsByHashtag(hashtag string) ([]Blogs, error) {
	var blogs []Blogs

	filter := bson.M{"hashtags": bson.M{"$elemMatch": bson.M{"name": hashtag}}}
	cursor, err := conn.CollectionBlogs.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var blog Blogs
		if err := cursor.Decode(&blog); err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if blogs == nil {
		return nil, errors.New("blogs not found")
	}
	return blogs, nil
}

func (conn *Conn) GetRelatedHashtags(keyword string) ([]Hashtags, error) {
	var relatedHashtags []Hashtags

	filter := bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: keyword, Options: "i"}}}
	cursor, err := conn.CollectionHashtags.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var relatedHashtag Hashtags
		if err := cursor.Decode(&relatedHashtag); err != nil {
			return nil, err
		}
		relatedHashtags = append(relatedHashtags, relatedHashtag)
	}

	return relatedHashtags, nil
}
