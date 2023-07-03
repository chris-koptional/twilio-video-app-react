package handler

import (
	queue "server/queue"

	t "server/twilio"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
)

func SetupRouter(taskClient *cloudtasks.Client, videoClient *twilio.RestClient) *gin.Engine {

	router := gin.Default()

	// defer client.Close()

	router.Use(queue.Inject_gcp_client(taskClient))
	router.Use(t.InjectTwilioClient(videoClient))
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

	return router
}
