package api

import (
	"github.com/gin-gonic/gin"
)

func Data(c *gin.Context) (any, error) {
	return gin.H{
		"items": []gin.H{
			{
				"id":   1,
				"name": "Item 1",
			},
			{
				"id":   2,
				"name": "Item 2",
			},
		},
		"total": 2,
	}, nil
}
