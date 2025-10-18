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
	Handler     func(c *gin.Context) (props any, err error)
	embedFS     *embed.FS
	port        string
}

type ErrorHandler func(c *gin.Context, err error, page *Page)

type Options struct {
	Router           *gin.Engine
	EmbedFS          *embed.FS
	Title            string
	MetaTags         []MetaTag
	Links            []Link
	PagesDir         string
	Handlers         map[string]func(c *gin.Context) (any, error)
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
	PagesDir string
	Handlers map[string]func(c *gin.Context) (any, error)
}
