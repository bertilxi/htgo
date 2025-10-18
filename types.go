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

type htmlTemplateData struct {
	RenderedContent template.HTML
	InitialProps    template.JS
	JS              template.JS
	CSS             template.CSS
	Title           template.HTML
	IsDev           bool
	Hydrate         bool
	RouteID         string
	MetaTags        []MetaTag
	Links           []Link
	Lang            template.HTML
	Class           template.HTML
	WebSocketPort   string
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
	Handler     func(c *gin.Context) Page
	embedFS     *embed.FS
	port        string
}

type Options struct {
	Router   *gin.Engine
	EmbedFS  *embed.FS
	Title    string
	MetaTags []MetaTag
	Links    []Link
	Pages    []Page
	Lang     string
	Class    string
	Port     string
}

type Engine struct {
	Options
}
