package handler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	openai "server/openAI"
	"server/twilio"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
)

// /api/tasks/createTranscription
func handleTranscribeRecording(c *gin.Context) {
	gcp_storage, ok := c.Get("storage_client")

	if !ok {
		fmt.Println("Could not get storage client")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "Could not get storage client",
		})
		return
	}

	storageClient, ok := gcp_storage.(*storage.Client)

	if !ok {
		fmt.Println("Could not get storage client")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "Storage client error",
		})
		return
	}

	var payload CompositionCallback

	if err := c.ShouldBind(&payload); err != nil {
		// Handle error
		fmt.Println("Could not bind payload")
		return
	}

	recordingId := "CJa2813a32eebcc2241aca846e82ddd319"
	// fetch video
	fileName, err := twilio.DownloadRecording(recordingId)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "Failed to download video",
		})
		return
	}
	defer twilio.DeleteLocalRecording(fileName)

	fileName, err = twilio.ConvertVideoToAudio(recordingId)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "failed to convert to audio file",
		})
		return
	}
	defer twilio.DeleteLocalRecording(fileName)

	segments, err := twilio.IsWithinSizeLimit(recordingId)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "Failed to check size limit",
		})
		return
	}
	var completeTranscription string
	if segments > 0 {
		err = twilio.SplitRecording(recordingId)

		if err != nil {
			c.JSON(http.StatusFailedDependency, gin.H{
				"error": "Failed splitting videos",
			})
			return
		}

		var wg sync.WaitGroup
		wg.Add(int(segments + 1))
		// c := make(chan string, segments+1)
		results := make([]string, segments+1)

		for i := 0; i < int(segments)+1; i++ {
			go func(i int) {

				fileName := fmt.Sprintf("recording-%s_00%d.mp3", recordingId, i)
				fmt.Println("Getting transcription ", i)
				transcription, _ := openai.GetVideoTranscription(fileName)

				results[i] = transcription
				twilio.DeleteLocalRecording(fileName)
				wg.Done()
			}(i)
		}
		wg.Wait()
		for _, transcript := range results {
			completeTranscription += transcript
		}
	} else {
		filename := fmt.Sprintf("recording-%s.mp3", recordingId)
		completeTranscription, _ = openai.GetVideoTranscription(filename)

		err = twilio.DeleteLocalRecording(recordingId)
	}

	// if err != nil {
	// 	fmt.Println("Failed to delete file after transcription")
	// }

	wc := storageClient.Bucket("knotion_transcriptions").Object(fmt.Sprintf("%s.txt", recordingId)).NewWriter(c)

	fileContent := []byte(completeTranscription)
	fileReader := ioutil.NopCloser(bytes.NewReader(fileContent))

	if _, err = io.Copy(wc, fileReader); err != nil {
		log.Fatalf("Failed to copy file content to writer: %v", err)
	}
	if err := wc.Close(); err != nil {
		log.Fatalf("Failed to close writer: %v", err)
	}
	c.JSON(http.StatusAccepted, gin.H{
		"transcription": "Stored in cloud",
	})
}
