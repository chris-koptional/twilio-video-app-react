package handler

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {

	router := gin.Default()

	// serve the react application
	router.Use(static.Serve("/", static.LocalFile("./build", true)))
	// Serve index.html for all non-API routes
	router.NoRoute(func(c *gin.Context) {
		c.File("./build/index.html")
	})

	// Twilio API
	router.POST("/api/room/create", handleTwilioRoomCreation)
	router.POST("/api/room/connect", handleTwilioTokenCreation)

	return router
}
