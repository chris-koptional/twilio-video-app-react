package twilio

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client/jwt"
	openapi "github.com/twilio/twilio-go/rest/video/v1"
)

const BYTES_PER_MB int64 = 1_048_576
const MAX_FILE_SIZE int64 = BYTES_PER_MB * 20 // 20MB

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
	} else {
		fmt.Println("Created Composition Hook")
	}

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
	callbackURL := fmt.Sprintf("%s/api/status/composition", url)
	fmt.Println("Callback URL:", callbackURL)
	hookParams.
		SetFriendlyName("Adhoc Meeting Compliation").
		SetVideoLayout(VideoLayout{
			Grid: Grid{
				VideoSources: []string{"*"},
			},
		}).
		SetStatusCallback(callbackURL).
		SetStatusCallbackMethod("POST").
		SetAudioSources([]string{"*"}).
		SetFormat("mp4")

	c.VideoV1.CreateCompositionHook(hookParams)

	return nil
}

func ConvertVideoToAudio(recordingId string) (string, error) {

	commandString := "ffmpeg -i recording-%s.mp4 recording-%s.mp3"

	commandString = fmt.Sprintf(commandString, recordingId, recordingId)

	args := strings.Split(commandString, " ")

	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()

	if err != nil {
		return "", err
	}
	return fmt.Sprintf("recording-%s.mp3", recordingId), nil
}

func DownloadRecording(id string) (string, error) {

	recordingURL := fmt.Sprintf("https://video.twilio.com/v1/Compositions/%s/Media", id)

	client := &http.Client{}

	req, err := http.NewRequest("GET", recordingURL, nil)
	if err != nil {
		fmt.Println("Failed to create request:", err)
		return "", err
	}
	accountSid := os.Getenv("TWILIO_ACCOUNT_ID")
	authToken := os.Getenv("TWILIO_AUTH_KEY")

	req.SetBasicAuth(accountSid, authToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return "", err
	}
	defer resp.Body.Close()

	fileName := fmt.Sprintf("recording-%s.mp4", id)

	file, err := os.Create(fileName)
	defer file.Close()

	if err != nil {
		fmt.Println("Failed to create file:", err)
		return "", err
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Failed to save recording:", err)
		file.Close()
		return "", err
	}

	return fileName, nil
}

func DeleteLocalRecording(fileName string) error {
	err := os.Remove(fileName)

	if err != nil {
		return err
	}
	return nil
}

func IsWithinSizeLimit(id string) (int64, error) {
	fileName := fmt.Sprintf("recording-%s.mp3", id)

	file, err := os.Stat(fileName)

	if err != nil {
		return 0, err
	}

	fileSize := file.Size()

	return fileSize / MAX_FILE_SIZE, nil
}

func SplitRecording(id string) error {

	commandString := "ffmpeg -i recording-%s.mp3 -c copy -f segment -segment_time 1200 -reset_timestamps 1 recording-%s_%%03d.mp3"

	commandString = fmt.Sprintf(commandString, id, id)
	fmt.Println(commandString)
	args := strings.Split(commandString, " ")

	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}
	return nil

}

func GetTwilioClient(c *gin.Context) (*twilio.RestClient, error) {

	val, ok := c.Get("video_client")

	if !ok {
		fmt.Println("Failed to get video client")
		return nil, errors.New("Failed to get video client.")
	}

	client, ok := val.(*twilio.RestClient)

	if !ok {
		fmt.Println("Improper video client")
		return nil, errors.New("Improper video client.")
	}
	return client, nil
}

func GetUniqueRoomName(c *twilio.RestClient, sid string) (string, error) {
	room, err := c.VideoV1.FetchRoom(sid)

	if err != nil {
		return "", err
	}

	return *room.UniqueName, nil
}
