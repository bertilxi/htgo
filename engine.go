package htgo

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

func (engine *Engine) HandleRoutes() {
	pages, err := DiscoverPages(engine.PagesDir, engine.Handlers)
	if err != nil {
		fmt.Printf("Error discovering pages: %v\n", err)
		os.Exit(1)
	}

	engine.Pages = pages

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
			Handlers:         options.Handlers,
			ErrorHandler:     options.ErrorHandler,
			AssetURLPrefix:   options.AssetURLPrefix,
			CacheBustVersion: options.CacheBustVersion,
		},
		PagesDir: options.PagesDir,
		Handlers: options.Handlers,
	}

	return engine
}
