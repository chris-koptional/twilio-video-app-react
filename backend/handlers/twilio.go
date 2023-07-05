package handler

import (
	"fmt"
	"net/http"
	"os"
	queue "server/cloud"
	"server/twilio"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/gin-gonic/gin"
	t "github.com/twilio/twilio-go"
)

type User struct {
	RoomID string `json:"roomId"`
	Name   string `json:"user"`
}

type RoomRequest struct {
	ID string `json:"documentId"`
}

func handleTwilioRoomCreation(c *gin.Context) {
	var room RoomRequest

	client, ok := c.Get("video_client")

	if !ok {
		c.JSON(http.StatusFailedDependency, gin.H{"error": "Could not source twilio client"})
		return
	}

	twilioClient, ok := client.(*t.RestClient)

	if !ok {
		c.JSON(http.StatusFailedDependency, gin.H{"error": "Twilio client improper initialization"})
		return
	}
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := twilio.CreateRoom(twilioClient, room.ID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create Room",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Created Room",
		"roomName": room.ID,
	})
}

func handleTwilioTokenCreation(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	token, err := twilio.CreateToken(user.RoomID, user.Name)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create JWT",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Created JWTet for Room",
		"token":   token,
	})
}

type CompositionCallback struct {
	CompositionSid      string `json:"CompositionSid,omitempty"`
	StatusCallbackEvent string `json:"StatusCallbackEvent,omitempty"`
}

func handleCompositionCallback(c *gin.Context) {

	var payload CompositionCallback

	if err := c.ShouldBind(&payload); err != nil {
		// Handle error
		fmt.Println("Could not bind payload")
	}

	if payload.StatusCallbackEvent != "composition-available" {
		c.JSON(http.StatusOK, gin.H{
			"message": "Callback received",
		})
		return
	}
	fmt.Println(payload, payload.CompositionSid, "COMPOSITION SID")

	client, ok := c.Get("task_client")

	if !ok {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "GCP client not found.",
		})
		return
	}

	gcpClient, ok := client.(*cloudtasks.Client)

	if !ok {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "GCP client not correct type.",
		})
		return
	}

	baseURL := os.Getenv("FLY_APP_DOMAIN")

	url := fmt.Sprintf("%s/api/tasks/createTranscription", baseURL)

	queue.CreatePostRecordingTask(gcpClient, url, payload)

	c.JSON(http.StatusOK, gin.H{
		"message": "Callback received",
	})
}
