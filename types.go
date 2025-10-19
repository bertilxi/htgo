package alloy

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

type Engine struct {
	Options
	Pages    []Page
	Loaders  map[string]PageLoader
	Handlers map[string]gin.HandlerFunc
}
