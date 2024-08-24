package graphQL

import (
	"be/controllers"
	"be/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query: graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"likeOrDislikeBlog": &graphql.Field{
				Type: graphql.NewObject(graphql.ObjectConfig{
					Name: "LikeOrDislike",
					Fields: graphql.Fields{
						"type": &graphql.Field{
							Type: graphql.String,
						},
						"userID": &graphql.Field{
							Type: graphql.ID,
						},
						"blogID": &graphql.Field{
							Type: graphql.ID,
						},
					},
				}),
				Args: graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"userID": &graphql.ArgumentConfig{
						Type: graphql.ID,
					},
					"blogID": &graphql.ArgumentConfig{
						Type: graphql.ID,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					fmt.Println("asdasd")
					model := models.NewConn()
					fmt.Println(">>>>", p.Args["userID"].(primitive.ObjectID), p.Args["blogID"].(primitive.ObjectID))

					data := controllers.ActionLikeOrDisLike{
						Type:   p.Args["type"].(string),
						UserID: p.Args["userID"].(primitive.ObjectID),
						BlogID: p.Args["blogID"].(primitive.ObjectID),
					}

					var err error
					if data.Type == "like" {
						err = model.LikeBlog(data.UserID, data.BlogID)
					} else {
						err = model.DisLikeBlog(data.UserID, data.BlogID)
					}

					if err != nil {
						return nil, err
					}

					return data, nil
				},
			},
		},
	}),
	Mutation: graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"likeOrDislikeBlog": &graphql.Field{
				Type: graphql.NewObject(graphql.ObjectConfig{
					Name: "LikeOrDislike",
					Fields: graphql.Fields{
						"type": &graphql.Field{
							Type: graphql.String,
						},
						"userID": &graphql.Field{
							Type: graphql.ID,
						},
						"blogID": &graphql.Field{
							Type: graphql.ID,
						},
					},
				}),
				Args: graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"userID": &graphql.ArgumentConfig{
						Type: graphql.ID,
					},
					"blogID": &graphql.ArgumentConfig{
						Type: graphql.ID,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					fmt.Println("asdasd")
					model := models.NewConn()
					userIDObj, _ := primitive.ObjectIDFromHex(p.Args["userID"].(string))
					blogIDObj, _ := primitive.ObjectIDFromHex(p.Args["blogID"].(string))
					fmt.Println(">>>>", p.Args["userID"])

					now := time.Now()
					data := controllers.ActionLikeOrDisLike{
						Type:       p.Args["type"].(string),
						UserID:     userIDObj,
						BlogID:     blogIDObj,
						Created_At: &now,
					}

					var err error
					if data.Type == "like" {
						err = model.LikeBlog(data.UserID, data.BlogID)
					} else {
						err = model.DisLikeBlog(data.UserID, data.BlogID)
					}

					if err != nil {
						return nil, err
					}

					return data, nil
				},
			},
		}}),
})

type GraphQLRequest struct {
	Query string `json:"query"`
}

func ActionLikeOrDislike(c *gin.Context) {
	var req GraphQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No query received"})
		return
	}

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: req.Query,
	})

	c.JSON(http.StatusOK, result)
}
