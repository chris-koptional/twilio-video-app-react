package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"server/cloud"
	t "server/twilio"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/storage"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
)

func SetupRouter(taskClient *cloudtasks.Client, videoClient *twilio.RestClient, storageClient *storage.Client) *gin.Engine {

	router := gin.Default()

	// defer client.Close()

	router.Use(cloud.Inject_gcp_client(taskClient))
	router.Use(t.InjectTwilioClient(videoClient))
	router.Use(cloud.InjectStorageClient(storageClient))
	// serve the react application
	router.Use(static.Serve("/", static.LocalFile("./build", true)))
	// Serve index.html for all non-API routes
	router.NoRoute(func(c *gin.Context) {
		c.File("./build/index.html")
	})

	// Twilio API
	router.POST("/api/room/create", handleTwilioRoomCreation)
	router.POST("/api/room/connect", handleTwilioTokenCreation)

	router.POST("/api/status/composition", handleCompositionCallback)

	// tasks
	router.POST("/api/tasks/createTranscription", handleTranscribeRecording)
	router.POST("/api/tasks/getSummary", handleSummarizeTranscript)
	router.POST("/api/tasks/updateSummary", func(c *gin.Context) {
		body, _ := ioutil.ReadAll(c.Request.Body)

		fmt.Println(string(body))

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "not implemented yet",
		})
	})
	return router
}
