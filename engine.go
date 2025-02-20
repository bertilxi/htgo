package htgo

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

func (engine *Engine) HandleRoutes() {
	for _, page := range engine.Pages {
		engine.Router.GET(page.Route, page.render)
	}

	engine.Router.Run()
}

func (engine *Engine) Start() {
	if engine.EmbedFS == nil {
		engine.Router.Static(CacheDir, CacheDir)
	} else {
		engine.Router.Any(CacheDir+"/*path", func(c *gin.Context) {
			route := c.Param("path")

			c.FileFromFS(path.Join(CacheDir, route), http.FS(engine.EmbedFS))
		})
	}

	engine.HandleRoutes()
}

func setupPages(options Options) []Page {
	appPages := []Page{}

	for _, page := range options.Pages {
		page.AssignOptions(options)

		appPages = append(appPages, page)
	}

	return appPages
}

func New(options Options) *Engine {
	if IsProd() || IsBuild() {
		gin.SetMode(gin.ReleaseMode)
	}

	if options.Router == nil {
		options.Router = gin.Default()
	}

	engine := &Engine{
		Options: Options{
			Router:   options.Router,
			EmbedFS:  options.EmbedFS,
			Title:    options.Title,
			MetaTags: options.MetaTags,
			Links:    options.Links,
			Lang:     options.Lang,
			Class:    options.Class,
			Pages:    setupPages(options),
		},
	}

	return engine
}
