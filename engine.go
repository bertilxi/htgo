package htgo

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

func (engine *Engine) HandleRoutes() {
	for _, page := range engine.Pages {
		engine.Router.GET(page.Route, page.render)
	}

	port := engine.Port
	if port == "" {
		port = "8080"
	}

	if IsDev() {
		fmt.Println()
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("✓ HTGO Dev Server Ready")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("🌐 Local:       http://localhost:%s\n", port)
		fmt.Println()
		fmt.Println("📄 Routes:")
		for _, page := range engine.Pages {
			fmt.Printf("   • %s\n", page.Route)
		}
		fmt.Println()
		fmt.Println("🔄 Hot reload enabled - changes will auto-refresh")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
	}

	engine.Router.Run(":" + port)
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

	port := options.Port
	if port == "" {
		port = "8080"
	}

	for _, page := range options.Pages {
		page.AssignOptions(options)
		page.port = port

		appPages = append(appPages, page)
	}

	return appPages
}

func New(options Options) *Engine {
	if IsProd() {
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
