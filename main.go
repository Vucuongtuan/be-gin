package main

import (
	"be/config"
	"be/routes"
	"context"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	client, err := config.ConnectionDB()
    if err != nil {
        fmt.Println(err)	
    }
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			fmt.Println("Failed to disconnect from MongoDB: %v", err)
		}
	}()



	// config res api
	r := gin.Default()
	//cros config
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	
	// r.Use(cors.Default())
	//socket.io


	

	// server := socketio.NewServer(nil)

	// server.OnConnect("/", func(s socketio.Conn) error {
	// 	s.SetContext("")
	// 	fmt.Println("connected:", s.ID())
	// 	return nil
	// })

	// server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
	// 	fmt.Println("notice:", msg)
	// 	s.Emit("reply", "have "+msg)
	// })

	// server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
	// 	s.SetContext(msg)
	// 	return "recv " + msg
	// })

	// server.OnEvent("/", "bye", func(s socketio.Conn) string {
	// 	last := s.Context().(string)
	// 	s.Emit("bye", last)
	// 	s.Close()
	// 	return last
	// })

	// server.OnError("/", func(s socketio.Conn, e error) {
	// 	fmt.Println("meet error:", e)
	// })

	// server.OnDisconnect("/", func(s socketio.Conn, reason string) {
	// 	fmt.Println("closed", reason)
	// })

	// go func() {
	// 	if err := server.Serve(); err != nil {
	// 		fmt.Println("socketio listen error: %s\n", err)
	// 	}
	// }()
	// defer server.Close()

	// r.GET("/socket.io/*any", gin.WrapH(server))
	// r.POST("/socket.io/*any", gin.WrapH(server))
	//api http
	api := r.Group("api")
	{
		v1 := api.Group("v1")
		{
			routes.SetupRoutes(v1)
		}
	}

	//socket 
	
	uploads := r.Group("/uploads")
	{
		uploads.Static("/image", "./uploads/image")
		uploads.Static("/audio", "./uploads/audio")
		uploads.Static("/video", "./uploads/video")
		uploads.Static("/othor", "./uploads")
	}
	//socket 



	//config post
	r.Run(":4000")
}


// func getUser(c *gin.Context) {
// 	id := c.Query("id")
// 	year := c.Query("year")
// 	 c.JSON(200, gin.H{
// 		"msg":"This is api method GET :" + year + id,
// 	})
// }

// func postUser(c *gin.Context) {
// 	var user TData
        
// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{
// 		"id":      user.Id,
// 		"name":    user.Name,
// 		"address": user.Address,
// 	})
// }
// func delete(c *gin.Context) {
// 	c.JSON(200, gin.H{
// 		"msg": "This is api method DELETE",
// 	})
// }
// func updateUser(c *gin.Context) {
// 	c.JSON(200, gin.H{
// 		"msg":"THis is api method UPDATE",
// 	})
// }