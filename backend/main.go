package main

import (
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type User struct {
	RoomID string `json:"roomId"`
	Name   string `json:"user"`
}

type RoomRequest struct {
	ID string `json:"documentId"`
}

func main() {
	r := gin.Default()
	godotenv.Load()
	// Serve static files (React app)
	r.Use(static.Serve("/", static.LocalFile("./build", true)))

	r.POST("/api/room/connect", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, err := create_token(user.RoomID, user.Name)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Failed to create JWT",
				"error":   true,
			})
			return
		}
		// Process the user data
		// You can access the roomID and username using user.RoomID and user.Username respectively
		// Perform the desired actions with the user data, such as storing it in a database

		c.JSON(http.StatusOK, gin.H{
			"message": "Created JWT for Room",
			"token":   token,
		})
	})

	r.POST("/api/room/create", func(c *gin.Context) {
		var room RoomRequest
		if err := c.ShouldBindJSON(&room); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := create_room(room.ID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Failed to create Room",
				"error":   true,
			})
			return
		}
		// Process the user data
		// You can access the roomID and username using user.RoomID and user.Username respectively
		// Perform the desired actions with the user data, such as storing it in a database
		c.JSON(http.StatusOK, gin.H{
			"message":  "Created Room",
			"roomName": room.ID,
			"error":    false,
		})
	})
	// Serve index.html for all non-API routes
	r.NoRoute(func(c *gin.Context) {
		c.File("./build/index.html")
	})
	// Start the server
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
