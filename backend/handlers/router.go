package handler

import (
	"fmt"
	"io"
	"net/http"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/storage"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
)

type Handler struct {
	TaskClient    *cloudtasks.Client
	VideoClient   *twilio.RestClient
	StorageClient *storage.Client
}

func NewHandler(taskClient *cloudtasks.Client, videoClient *twilio.RestClient, storageClient *storage.Client) *Handler {
	return &Handler{
		TaskClient:    taskClient,
		VideoClient:   videoClient,
		StorageClient: storageClient,
	}
}

func (h *Handler) SetupRouter() *gin.Engine {

	router := gin.Default()

	// serve the react application
	router.Use(static.Serve("/", static.LocalFile("./build", true)))
	// Serve index.html for all non-API routes
	router.NoRoute(func(c *gin.Context) {
		c.File("./build/index.html")
	})

	// Twilio API
	router.POST("/api/room/create", h.handleTwilioRoomCreation)
	router.POST("/api/room/connect", h.handleTwilioTokenCreation)

	router.POST("/api/status/composition", h.handleCompositionCallback)

	// tasks
	router.POST("/api/tasks/createTranscription", h.handleTranscribeRecording)
	router.POST("/api/tasks/getSummary", h.handleSummarizeTranscript)
	router.POST("/api/tasks/updateSummary", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)

		fmt.Println(string(body))

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "not implemented yet",
		})
	})
	return router
}
