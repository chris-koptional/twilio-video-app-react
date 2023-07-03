package twilio

import (
	"errors"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client/jwt"
	openapi "github.com/twilio/twilio-go/rest/video/v1"
)

type VideoLayout struct {
	Grid Grid `json:"grid"`
}

type Grid struct {
	VideoSources []string `json:"video_sources"`
}

func InjectTwilioClient(client *twilio.RestClient) gin.HandlerFunc {

	err := CreateCompositionHook(client)

	if err != nil {
		fmt.Println("Failed to create composition hook!")
	}
	fmt.Println("Created Composition Hook")
	return func(c *gin.Context) {
		c.Set("video_client", client)
		c.Next()
	}
}

func InitTwilioClient() *twilio.RestClient {
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

func CreateToken(room, user string) (string, error) {
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

func CreateRoom(client *twilio.RestClient, roomId string) error {

	params := &openapi.CreateRoomParams{}

	params = params.
		SetType("group").
		SetRecordingRules(params.RecordParticipantsOnConnect).
		SetUniqueName(roomId)

	_, err := client.VideoV1.CreateRoom(params)

	if err != nil {
		return err
	}

	return nil
}

func CreateCompositionHook(c *twilio.RestClient) error {
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
