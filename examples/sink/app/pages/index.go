package pages

import (
	"time"

	"github.com/gin-gonic/gin"
)

// This is a loader for index.tsx
// The function name doesn't matter, just needs to match the signature:
// func(c *gin.Context) (any, error)
func LoadIndex(c *gin.Context) (any, error) {
	return map[string]any{
		"route": c.FullPath(),
		"time":  time.Now().String(),
	}, nil
}
