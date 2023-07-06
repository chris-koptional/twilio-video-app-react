package handler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"server/cloud"
	openai "server/openAI"
	"server/twilio"
	"sync"

	"github.com/gin-gonic/gin"
)

type TranscribeRecordingTaskPayload struct {
	ID string `json:"id"`
}

// /api/tasks/createTranscription
func handleTranscribeRecording(c *gin.Context) {

	storageClient, err := cloud.GetStorageClient(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get storage client.",
		})
		return
	}

	taskClient, err := cloud.GetTaskClient(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get task client.",
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
				transcription, e := openai.GetVideoTranscription(fileName)
				if e != nil {
					err = e
				}
				results[i] = transcription
				twilio.DeleteLocalRecording(fileName)
				wg.Done()
			}(i)
		}
		wg.Wait()

		if err != nil {
			c.JSON(http.StatusFailedDependency, gin.H{
				"error": err.Error(),
			})
			return
		}
		for _, transcript := range results {
			completeTranscription += transcript
		}
	} else {
		filename := fmt.Sprintf("recording-%s.mp3", recordingId)
		completeTranscription, _ = openai.GetVideoTranscription(filename)

		twilio.DeleteLocalRecording(filename)
	}

	wc := storageClient.Bucket("knotion_transcriptions").Object(fmt.Sprintf("%s.txt", recordingId)).NewWriter(c)

	fileContent := []byte(completeTranscription)
	fileReader := ioutil.NopCloser(bytes.NewReader(fileContent))

	if _, err = io.Copy(wc, fileReader); err != nil {
		log.Fatalf("Failed to copy file content to writer: %v", err)
	}
	if err := wc.Close(); err != nil {
		log.Fatalf("Failed to close writer: %v", err)
	}

	baseURL := os.Getenv("FLY_APP_DOMAIN")

	updateNotionTranscriptURL := fmt.Sprintf("%s/api/tasks/updateNotionTranscript", baseURL)
	getSummaryURL := fmt.Sprintf("%s/api/tasks/getSummary", baseURL)
	taskPayload := TranscribeRecordingTaskPayload{recordingId}

	_, err = cloud.CreatePostRecordingTask(taskClient, updateNotionTranscriptURL, taskPayload)
	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": err.Error(),
		})
		return
	}

	_, err = cloud.CreatePostRecordingTask(taskClient, getSummaryURL, taskPayload)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"transcription": "Stored in cloud",
	})
}
