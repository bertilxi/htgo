package pages

import (
	"github.com/gin-gonic/gin"
)

// Loader for about.tsx
func LoadAbout(c *gin.Context) (any, error) {
	return map[string]any{
		"description": "This is the about page",
	}, nil
}
