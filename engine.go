package alloy

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/bertilxi/alloy/core"
)

// RegisterRoutes registers all pages and handlers to the Gin router.
// Call this after creating the engine but before calling Start().
func (engine *Engine) RegisterRoutes() error {
	pages, err := DiscoverPages(engine.PagesDir, engine.Loaders)
	if err != nil {
		return fmt.Errorf("discover pages: %w", err)
	}

	engine.Pages = pages

	// Register API handlers first (so they take precedence over page routes)
	if engine.Handlers != nil && len(engine.Handlers) > 0 {
		for route, handler := range engine.Handlers {
			engine.Router.Any(route, handler)
		}
	}

	for i := range engine.Pages {
		engine.Pages[i].AssignOptions(engine.Options)
		engine.Router.GET(engine.Pages[i].Route, engine.Pages[i].Render)
	}

	return nil
}

// RegisterBundles registers the bundle static file handler.
// Serves bundles from disk (dev) or embedded FS (production).
func (engine *Engine) RegisterBundles() {
	if engine.EmbedFS == nil {
		engine.Router.Static(core.CacheDir, core.CacheDir)
	} else {
		engine.Router.Any(core.CacheDir+"/*path", func(c *gin.Context) {
			route := c.Param("path")
			c.FileFromFS(path.Join(core.CacheDir, route), http.FS(engine.EmbedFS))
		})
	}
}

// Listen starts the HTTP server on the configured port.
// Must call RegisterRoutes() and RegisterBundles() before this.
func (engine *Engine) Listen() error {
	port := engine.Port
	if port == "" {
		port = "8080"
	}
	return engine.Router.Run(":" + port)
}

// Start is a convenience method that calls RegisterBundles(), RegisterRoutes(), and Listen() in order.
// Deprecated: Use RegisterBundles(), RegisterRoutes(), and Listen() directly for more control.
func (engine *Engine) Start() error {
	engine.RegisterBundles()
	if err := engine.RegisterRoutes(); err != nil {
		return err
	}
	return engine.Listen()
}

// AssignOptions assigns global options to a page.
func (page *Page) AssignOptions(options Options) {
	page.embedFS = options.EmbedFS
	page.ErrorHandler = options.ErrorHandler
	page.Class = options.Class
	page.Links = append(page.Links, options.Links...)
	page.MetaTags = append(page.MetaTags, options.MetaTags...)
	page.Lang = options.Lang

	if page.Lang == "" {
		page.Lang = "en"
	}
	if page.Title == "" {
		page.Title = options.Title
	}
}

// assetURL constructs a proper asset URL path.
func (page *Page) assetURL(path string) string {
	url := "/" + path
	if strings.HasPrefix(url, "//") {
		url = strings.TrimPrefix(url, "/")
	}
	return url
}

func New(options Options) *Engine {
	if core.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	if options.Router == nil {
		options.Router = gin.Default()
	}

	port := options.Port
	if port == "" {
		port = os.Getenv("PORT")
	}

	pagesDir := options.PagesDir
	if pagesDir == "" {
		pagesDir = "./pages"
	}

	engine := &Engine{
		Options: Options{
			Router:       options.Router,
			EmbedFS:      options.EmbedFS,
			Title:        options.Title,
			MetaTags:     options.MetaTags,
			Links:        options.Links,
			Lang:         options.Lang,
			Class:        options.Class,
			Port:         port,
			PagesDir:     pagesDir,
			Loaders:      options.Loaders,
			Handlers:     options.Handlers,
			ErrorHandler: options.ErrorHandler,
		},
		Loaders:  options.Loaders,
		Handlers: options.Handlers,
	}

	return engine
}
