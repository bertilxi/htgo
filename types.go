package htgo

import (
	"embed"
	"html/template"

	"github.com/gin-gonic/gin"
)

type MetaTag struct {
	Name     template.HTML
	Content  template.HTML
	Property template.HTML
}

type Link struct {
	Rel  template.HTML
	Href template.HTML
}

type Page struct {
	Route       string
	File        string
	Interactive bool
	Props       any
	Title       string
	MetaTags    []MetaTag
	Links       []Link
	Lang        string
	Class       string
	Loader      PageLoader
	embedFS     *embed.FS
	port        string
}

type ErrorHandler func(c *gin.Context, err error, page *Page)

// PageLoader loads data for a page's SSR, returning props for the React component.
// Signature: func(c *gin.Context) (props any, err error)
type PageLoader func(c *gin.Context) (any, error)

// Handler handles an HTTP request with full Gin API control.
// Used for API endpoints and other non-page routes.
// Handlers can use c.JSON(), c.File(), c.String(), etc. directly.
// Return error if something went wrong; handler is responsible for setting response.
type Handler func(c *gin.Context) error

type Options struct {
	Router           *gin.Engine
	EmbedFS          *embed.FS
	Title            string
	MetaTags         []MetaTag
	Links            []Link
	PagesDir         string
	Loaders          map[string]PageLoader
	Handlers         map[string]Handler
	Lang              string
	Class            string
	Port             string
	ErrorHandler     ErrorHandler
	AssetURLPrefix   string
	CacheBustVersion string
}

type Engine struct {
	Options
	Pages    []Page
	Loaders  map[string]PageLoader
	Handlers map[string]Handler
}
