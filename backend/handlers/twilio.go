package handler

import (
	"fmt"
	"net/http"
	"os"
	queue "server/cloud"
	"server/twilio"

	"github.com/gin-gonic/gin"
)

type User struct {
	RoomID string `json:"roomId"`
	Name   string `json:"user"`
}

type RoomRequest struct {
	ID string `json:"documentId"`
}

func (h *Handler) handleTwilioRoomCreation(c *gin.Context) {
	var room RoomRequest

	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := twilio.CreateRoom(h.VideoClient, room.ID)

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

func (h *Handler) handleTwilioTokenCreation(c *gin.Context) {
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

func (h *Handler) handleCompositionCallback(c *gin.Context) {

	payload := CompositionCallback{
		c.Request.FormValue("CompositionSid"),
		c.Request.FormValue("StatusCallbackEvent"),
		c.Request.FormValue("RoomSid"),
	}

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

	roomName, err := twilio.GetUniqueRoomName(h.VideoClient, payload.RoomSID)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "Failed to get unique room name.",
		})
		return
	}

	baseURL := os.Getenv("FLY_APP_DOMAIN")

	url := fmt.Sprintf("%s/api/tasks/createTranscription", baseURL)

	taskPayload := CompositionTaskPayload{
		payload.CompositionSid,
		roomName,
	}
	queue.CreatePostRecordingTask(h.TaskClient, url, taskPayload)

	c.JSON(http.StatusOK, gin.H{
		"message": "Callback received",
	})
}
