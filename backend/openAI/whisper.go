package openai

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

type WhisperResponse struct {
	Text string `json:"text"`
}

func GetVideoTranscription(filePath string) (string, error) {

	token := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(token)

	resp, err := client.CreateTranscription(
		context.Background(),
		openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: filePath,
			Format:   openai.AudioResponseFormatText,
		},
	)

	if err != nil {
		fmt.Println("Failed to get transcription")
		return "", err
	}
	fmt.Println(resp)
	return resp.Text, nil
}

func GetTranscriptionSummary(transcript string) (string, error) {

	token := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(token)

	req := openai.ChatCompletionRequest{
		Model: "gpt-4-0613",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("I have a transcript from a video call recording from a business meeting. Could you please provide me with a summary in bullet form that highlights key information from the meeting. Here is the transcript: %s", transcript),
			},
		},
	}
	resp, err := client.CreateChatCompletion(context.Background(), req)

	if err != nil {
		fmt.Println("Failed to get transcription summary")
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
