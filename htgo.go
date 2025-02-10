package htgo

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

func New(config HtgoConfig) {
	config.Options = MapOptions(config.Options)

	if config.EmbedFS != nil {
		SetEmbedFS(config.EmbedFS)
	}

	if !IsDev() {
		config.Router.Any(CacheDir+"/*path", func(c *gin.Context) {
			route := c.Param("path")

			c.FileFromFS(path.Join(CacheDir, route), http.FS(config.EmbedFS))
		})
	}

	appPages := GetPages(config.Options)

	for _, page := range appPages {
		config.Router.GET(page.Route, config.Options.Mode.RenderPage(page))
	}
}
