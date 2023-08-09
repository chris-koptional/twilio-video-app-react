package handler

import (
	"fmt"
	"net/http"
	"os"
	"server/cloud"
	queue "server/cloud"
	"server/twilio"

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
	CompositionSid      string `form:"CompositionSid,omitempty"`
	StatusCallbackEvent string `form:"StatusCallbackEvent,omitempty"`
	RoomSID             string `form:"RoomSid,omitempty"`
}

type CompositionTaskPayload struct {
	CompositionSid string `json:"id,omitempty"`
	RoomName       string `json:"roomname"`
}

func handleCompositionCallback(c *gin.Context) {
	twilioClient, err := twilio.GetTwilioClient(c)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"errors": "Failed to get video rest client",
		})
		return
	}

	payload := CompositionCallback{
		c.Request.FormValue("CompositionSid"),
		c.Request.FormValue("StatusCallbackEvent"),
		c.Request.FormValue("RoomSid"),
	}

	fmt.Println(fmt.Sprintf("CompId:%s  Status: %s  RoomId: %s", payload.CompositionSid, payload.StatusCallbackEvent, payload.RoomSID))

	if err := c.ShouldBindQuery(&payload); err != nil {
		// Handle error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Improper payload",
		})
		fmt.Println("Could not bind payload")
	}

	if payload.StatusCallbackEvent != "composition-available" {
		c.JSON(http.StatusOK, gin.H{
			"message": "Callback received",
		})
		return
	}
	fmt.Println(payload, payload.CompositionSid, "COMPOSITION SID")

	roomName, err := twilio.GetUniqueRoomName(twilioClient, payload.RoomSID)

	taskClient, err := cloud.GetTaskClient(c)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "Could not get task client: ." + err.Error(),
		})
		return
	}

	baseURL := os.Getenv("FLY_APP_DOMAIN")

	url := fmt.Sprintf("%s/api/tasks/createTranscription", baseURL)

	taskPayload := CompositionTaskPayload{
		payload.CompositionSid,
		roomName,
	}
	queue.CreatePostRecordingTask(taskClient, url, taskPayload)

	c.JSON(http.StatusOK, gin.H{
		"message": "Callback received",
	})
}
