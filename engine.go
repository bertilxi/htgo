package htgo

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

func (engine *Engine) HandleRoutes() {
	pages, err := DiscoverPages(engine.PagesDir, engine.Loaders)
	if err != nil {
		fmt.Printf("Error discovering pages: %v\n", err)
		os.Exit(1)
	}

	engine.Pages = pages

	// Register API handlers first (so they take precedence over page routes)
	if engine.Handlers != nil && len(engine.Handlers) > 0 {
		for route, handler := range engine.Handlers {
			// Create a local copy to avoid closure issues
			handlerCopy := handler
			// Register handler with Any method (all HTTP verbs)
			engine.Router.Any(route, func(c *gin.Context) {
				err := handlerCopy(c)
				if err != nil && c.Writer.Status() == http.StatusOK {
					// Only set error response if handler didn't already write response
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}
			})
		}
	}

	for i := range engine.Pages {
		engine.Pages[i].AssignOptions(engine.Options)
		engine.Router.GET(engine.Pages[i].Route, engine.Pages[i].Render)
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

func New(options Options) *Engine {
	if IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	if options.Router == nil {
		options.Router = gin.Default()
	}

	port := options.Port
	if port == "" {
		port = os.Getenv("PORT")
	}

	engine := &Engine{
		Options: Options{
			Router:           options.Router,
			EmbedFS:          options.EmbedFS,
			Title:            options.Title,
			MetaTags:         options.MetaTags,
			Links:            options.Links,
			Lang:             options.Lang,
			Class:            options.Class,
			Port:             port,
			PagesDir:         options.PagesDir,
			Loaders:          options.Loaders,
			Handlers:         options.Handlers,
			ErrorHandler:     options.ErrorHandler,
			AssetURLPrefix:   options.AssetURLPrefix,
			CacheBustVersion: options.CacheBustVersion,
		},
		Loaders:  options.Loaders,
		Handlers: options.Handlers,
	}

	return engine
}
