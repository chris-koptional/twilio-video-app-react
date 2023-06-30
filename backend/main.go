package main

import (
	handler "server/handlers"

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
	godotenv.Load()

	// setup all of the routes
	r := handler.SetupRouter()

	// Start the server
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
