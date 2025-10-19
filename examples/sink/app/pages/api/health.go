package api

import (
	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) (any, error) {
	return gin.H{
		"status":  "ok",
		"message": "API is healthy",
	}, nil
}
