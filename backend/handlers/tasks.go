package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"server/cloud"
	openai "server/openAI"
	"server/twilio"
	"sync"

	"github.com/gin-gonic/gin"
)

type TranscribeRecordingTaskPayload struct {
	ID       string `json:"id"`
	RoomName string `json:"roomname"`
}
type GetSummaryTaskPayload struct {
	ID       string `json:"id"`
	RoomName string `json:"roomname"`
}

// /api/tasks/createTranscription
func (h *Handler) handleTranscribeRecording(c *gin.Context) {
	var payload CompositionTaskPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		// Handle error
		fmt.Println("Could not bind payload")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "Could not bind payload",
		})
		return
	}

	recordingId := payload.CompositionSid

	// fetch video
	fileName, err := twilio.DownloadRecording(recordingId)

	if err != nil {
		fmt.Println("Failed to download video")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "Failed to download video",
		})
		return
	}
	defer twilio.DeleteLocalRecording(fileName)

	fileName, err = twilio.ConvertVideoToAudio(recordingId)

	if err != nil {
		fmt.Println("Failed to convert to audio file", err)
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "failed to convert to audio file: " + err.Error(),
		})
		return
	}
	defer twilio.DeleteLocalRecording(fileName)

	segments, err := twilio.IsWithinSizeLimit(recordingId)

	if err != nil {
		fmt.Println("Failed to check size limit")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": "Failed to check size limit",
		})
		return
	}
	var completeTranscription string
	if segments > 0 {
		err = twilio.SplitRecording(recordingId)

		if err != nil {
			fmt.Println("Failed to split video")
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
			fmt.Println("Failed to get transcription")
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

	wc := h.StorageClient.Bucket("knotion_transcriptions").Object(fmt.Sprintf("%s.txt", recordingId)).NewWriter(c)

	fileContent := []byte(completeTranscription)
	fileReader := io.NopCloser(bytes.NewReader(fileContent))

	if _, err = io.Copy(wc, fileReader); err != nil {
		fmt.Println("Failed to copy conent to writer")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := wc.Close(); err != nil {
		fmt.Println("Failed to close writer")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": err.Error(),
		})
		return
	}

	baseURL := os.Getenv("FLY_APP_DOMAIN")

	updateNotionTranscriptURL := fmt.Sprintf("%s/api/tasks/updateNotionTranscript", baseURL)
	getSummaryURL := fmt.Sprintf("%s/api/tasks/getSummary", baseURL)
	taskPayload := TranscribeRecordingTaskPayload{recordingId, payload.RoomName}

	_, err = cloud.CreatePostRecordingTask(h.TaskClient, updateNotionTranscriptURL, taskPayload)
	if err != nil {
		fmt.Println("Failed to update notion transcript task")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": err.Error(),
		})
		return
	}

	_, err = cloud.CreatePostRecordingTask(h.TaskClient, getSummaryURL, taskPayload)

	if err != nil {
		fmt.Println("Failed to get summary task")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"transcription": "Stored in cloud",
	})
}

func (h *Handler) handleSummarizeTranscript(c *gin.Context) {
	var payload TranscribeRecordingTaskPayload

	err := c.ShouldBindJSON(&payload)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Improper request payload.",
		})
		return
	}

	transcript, err := cloud.GetTranscriptionObject(h.StorageClient, payload.ID)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": fmt.Sprintf("Failed getting transcript: %s", err.Error()),
		})
		return
	}
	summary, err := openai.GetTranscriptionSummary(transcript)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": fmt.Sprintf("Failed getting summary: %s", err.Error()),
		})
		return
	}

	wc := h.StorageClient.Bucket("knotion_summaries").Object(fmt.Sprintf("%s.txt", payload.ID)).NewWriter(c)

	fileContent := []byte(summary)
	fileReader := io.NopCloser(bytes.NewReader(fileContent))

	if _, err = io.Copy(wc, fileReader); err != nil {
		fmt.Println("Failed to copy conent to writer")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := wc.Close(); err != nil {
		fmt.Println("Failed to close writer")
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": err.Error(),
		})
		return
	}

	baseURL := os.Getenv("FLY_APP_DOMAIN")
	taskURL := fmt.Sprintf("%s/api/tasks/updateSummary", baseURL)

	taskPayload := GetSummaryTaskPayload{
		ID:       payload.ID,
		RoomName: payload.RoomName,
	}
	_, err = cloud.CreatePostRecordingTask(h.TaskClient, taskURL, taskPayload)

	if err != nil {
		c.JSON(http.StatusFailedDependency, gin.H{
			"error": fmt.Sprintf("Failed to create update summary task: %s", err.Error()),
		})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Summary created and new task issued.",
	})
}
