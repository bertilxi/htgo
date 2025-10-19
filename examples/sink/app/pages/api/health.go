package api

import (
	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) error {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "API is healthy",
	})
	return nil
}
