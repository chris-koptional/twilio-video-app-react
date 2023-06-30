package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client/jwt"
	openapi "github.com/twilio/twilio-go/rest/video/v1"
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

	c.JSON(http.StatusOK, gin.H{
		"message":  "Created Room",
		"roomName": room.ID,
		"error":    false,
	})
}

func handleTwilioTokenCreation(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"error":   true,
		})
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

	c.JSON(http.StatusOK, gin.H{
		"message": "Created JWT for Room",
		"token":   token,
		"error":   false,
	})
}

func handleCompositionCallback(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Callback received",
	})
}

func init_twilio_client() *twilio.RestClient {
	accountSid := os.Getenv("TWILIO_ACCOUNT_ID")
	twilioAPIKey := os.Getenv("TWILIO_API_KEY")
	apiSecret := os.Getenv("TWILIO_API_SECRET")

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username:   twilioAPIKey,
		Password:   apiSecret,
		AccountSid: accountSid,
	})

	return client
}

func create_token(room, user string) (string, error) {
	accountSid := os.Getenv("TWILIO_ACCOUNT_ID")
	twilioAPIKey := os.Getenv("TWILIO_API_KEY")
	apiSecret := os.Getenv("TWILIO_API_SECRET")

	params := jwt.AccessTokenParams{
		AccountSid:    accountSid,
		SigningKeySid: twilioAPIKey,
		Secret:        apiSecret,
		Identity:      user,
	}

	jwtToken := jwt.CreateAccessToken(params)

	videoGrant := &jwt.VideoGrant{
		Room: room,
	}
	jwtToken.AddGrant(videoGrant)
	token, err := jwtToken.ToJwt()

	if err != nil {
		error := fmt.Errorf("error: %q", err)
		fmt.Println(error.Error())
		return "", err
	}
	return token, nil
}

func create_room(roomId string) error {
	client := init_twilio_client()
	err := create_composition_hook(client)

	if err != nil {
		fmt.Println("Failed to create composition hook!")
	}
	params := &openapi.CreateRoomParams{}

	params = params.SetType("group").SetRecordingRules(params.RecordParticipantsOnConnect).SetUniqueName(roomId)
	_, err = client.VideoV1.CreateRoom(params)

	if err != nil {
		return err
	}

	return nil
}

type VideoLayout struct {
	Grid Grid `json:"grid"`
}

type Grid struct {
	VideoSources []string `json:"video_sources"`
}

func create_composition_hook(c *twilio.RestClient) error {
	url := os.Getenv("FLY_APP_DOMAIN")
	if url == "" {
		return errors.New("Could not source domain URL")
	}
	hookParams := &openapi.CreateCompositionHookParams{}

	hookParams.
		SetFriendlyName("Adhoc Meeting Compliation").
		SetVideoLayout(VideoLayout{
			Grid: Grid{
				VideoSources: []string{"*"},
			},
		}).
		SetStatusCallback(fmt.Sprintf("%s/api/status/composition", url)).
		SetStatusCallbackMethod("POST").
		SetAudioSources([]string{"*"}).
		SetFormat("mp4")

	c.VideoV1.CreateCompositionHook(hookParams)

	return nil
}
