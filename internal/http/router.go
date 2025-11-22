package http

import (
	"github.com/gin-gonic/gin"

	"github.com/freikugel0/boorumesh-be/internal/http/handler"
)

func NewRouter(devSrcHandler *handler.DevSourceHandler, apiHandler *handler.ApiHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	dev := r.Group("/dev")
	{
		dev.POST("/sources", devSrcHandler.Create)
		dev.GET("/sources/:code", devSrcHandler.GetSourceByCode)
	}

	api := r.Group("/api")
	{
		api.GET("/:source", apiHandler.GetImagesBySource)
	}

	return r
}
