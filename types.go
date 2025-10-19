package alloy

import (
	"embed"
	"html/template"

	"github.com/gin-gonic/gin"
)

// MetaTag defines SEO metadata for pages. Values are sanitized for HTML templates.
type MetaTag struct {
	Name     template.HTML
	Content  template.HTML
	Property template.HTML
}

// Link defines head links for pages. Values are sanitized for HTML templates.
type Link struct {
	Rel  template.HTML
	Href template.HTML
}

// Page defines a single route with its component and metadata.
type Page struct {
	Route        string
	File         string
	Interactive  bool
	Props        any
	Title        string
	MetaTags     []MetaTag
	Links        []Link
	Lang         string
	Class        string
	Loader       PageLoader
	ErrorHandler ErrorHandler
	embedFS      *embed.FS
}

// ErrorHandler is a framework-specific callback for rendering errors.
// Receives Gin context for framework-specific handling.
type ErrorHandler func(c *gin.Context, err error, page *Page)

// PageLoader loads data for a page's SSR, returning props for the React component.
// Signature: func(c *gin.Context) (props any, err error)
type PageLoader func(c *gin.Context) (any, error)

// Options configures the Alloy engine.
type Options struct {
	Router       *gin.Engine
	EmbedFS      *embed.FS
	Title        string
	MetaTags     []MetaTag
	Links        []Link
	PagesDir     string
	Loaders      map[string]PageLoader
	Handlers     map[string]gin.HandlerFunc
	Lang         string
	Class        string
	Port         string
	ErrorHandler ErrorHandler
}

// Engine manages routing, page discovery, and rendering.
type Engine struct {
	Options
	Pages    []Page
	Loaders  map[string]PageLoader
	Handlers map[string]gin.HandlerFunc
}
