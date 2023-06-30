package main

import (
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client/jwt"
	openapi "github.com/twilio/twilio-go/rest/video/v1"
)

func init_twilio_client() *twilio.RestClient {
	accountSid := os.Getenv("TWILIO_ACCOUNT_ID")
	twilioAPIKey := os.Getenv("TWILIO_API_KEY")
	apiSecret := os.Getenv("TWILIO_API_SECRET")

	return twilio.NewRestClientWithParams(twilio.ClientParams{
		Username:   twilioAPIKey,
		Password:   apiSecret,
		AccountSid: accountSid,
	})
}

func create_token(room, user string) (string, error) {
	accountSid := os.Getenv("TWILIO_ACCOUNT_ID")
	twilioAPIKey := os.Getenv("TWILIO_API_KEY")
	apiSecret := os.Getenv("TWILIO_API_SECRET")

	println(twilioAPIKey, "KEY HERE")
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
	params := &openapi.CreateRoomParams{}

	params = params.SetType("group").SetRecordingRules(params.RecordParticipantsOnConnect).SetUniqueName(roomId)
	_, err := client.VideoV1.CreateRoom(params)

	if err != nil {
		return err
	}

	return nil
}
