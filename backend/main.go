package main

import (
	"log"
	handler "server/handlers"
	"server/queue"
	"server/twilio"

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
	taskClient, err := queue.CreateClient()

	if err != nil {
		log.Fatalln("Failed to create GCP Cloud task client")
	}
	defer taskClient.Close()

	twilioClient := twilio.InitTwilioClient()

	// setup all of the routes
	r := handler.SetupRouter(
		taskClient,
		twilioClient,
	)

	// Start the server
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
