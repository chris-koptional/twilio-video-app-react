package main

import (
	"log"
	"server/cloud"
	handler "server/handlers"
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
	taskClient, err := cloud.CreateClient()

	if err != nil {
		log.Println("Failed to create GCP Cloud task client")
	}
	defer taskClient.Close()

	twilioClient := twilio.InitTwilioClient()

	storageClient, err := cloud.CreateStorageClient()

	if err != nil {
		log.Println("Failed to create GCP Cloud task client")
	}
	defer storageClient.Close()

	// setup all of the routes
	r := handler.SetupRouter(
		taskClient,
		twilioClient,
		storageClient,
	)

	// Start the server
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
