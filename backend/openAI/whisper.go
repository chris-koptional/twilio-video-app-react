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
	// url := "https://api.openai.com/v1/audio/transcriptions"

	// model := "whisper-1"

	// client := &http.Client{}

	// // Create a new HTTP request
	// request, err := http.NewRequest("POST", url, file)
	// if err != nil {
	// 	fmt.Println("Failed to create request:", err)
	// 	return "", err
	// }

	// // Set the authorization header
	// request.Header.Set("Authorization", "Bearer "+token)

	// // Create a multipart/form-data writer
	// body := &bytes.Buffer{}
	// writer := multipart.NewWriter(body)

	// // Create a form file field
	// fileField, err := writer.CreateFormFile("file", filePath)
	// if err != nil {
	// 	fmt.Println("Failed to create form file field:", err)
	// 	return "", err
	// }

	// // Copy the file contents to the form file field
	// _, err = io.Copy(fileField, file)
	// if err != nil {
	// 	fmt.Println("Failed to copy file contents:", err)
	// 	return "", err
	// }

	// // Add other form fields
	// _ = writer.WriteField("model", model)

	// // Close the multipart writer
	// err = writer.Close()
	// if err != nil {
	// 	fmt.Println("Failed to close multipart writer:", err)
	// 	return "", err
	// }

	// // Set the request body and content type
	// request.Body = ioutil.NopCloser(body)
	// request.Header.Set("Content-Type", writer.FormDataContentType())

	// // Send the HTTP request

	// // Print the request details
	// fmt.Println("Request URL:", request.URL.String())
	// fmt.Println("Request Method:", request.Method)
	// fmt.Println("Request Headers:")
	// for key, values := range request.Header {
	// 	for _, value := range values {
	// 		fmt.Printf("%s: %s\n", key, value)
	// 	}
	// }
	// fmt.Println("Request Body:")
	// bodyBytes, err := ioutil.ReadAll(request.Body)
	// if err != nil {
	// 	fmt.Println("Failed to read request body:", err)

	// }
	// fmt.Println(string(bodyBytes))

	// response, err := client.Do(request)
	// if err != nil {
	// 	fmt.Println("Failed to send request:", err)
	// 	return "", err
	// }
	// defer response.Body.Close()

	// // Read the response body
	// responseBody, err := ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	fmt.Println("Failed to read response body:", err)
	// 	return "", err
	// }
	// var whisperResponse WhisperResponse
	// fmt.Println(string(responseBody))
	// err = json.Unmarshal(responseBody, &whisperResponse)

	// if err != nil {
	// 	fmt.Println("Failed to marshal response body:", err.Error())
	// 	return "", err
	// }

	// return whisperResponse.Text, nil
}
