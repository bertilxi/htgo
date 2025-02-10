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

type HtmlTemplateData struct {
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
}

type pageMode string

const (
	PageModeJS   pageMode = "js"
	PageModeNoJS pageMode = "nojs"
)

type Page struct {
	Route    string
	File     string
	Mode     pageMode
	Props    any
	Title    string
	MetaTags []MetaTag
	Links    []Link
	Lang     string
	Class    string
	Handler  func(c *gin.Context) Page
}

type mode string

const (
	ModeStatic mode = "static"
	ModeSSR    mode = "ssr"
)

type HtgoMode struct {
	Name       mode
	RenderPage func(page Page) func(c *gin.Context)
}

var HtgoModeStatic = HtgoMode{
	Name:       ModeStatic,
	RenderPage: renderPage,
}

type SetupOptions struct {
	Mode     *HtgoMode
	Title    string
	MetaTags []MetaTag
	Links    []Link
	Pages    []Page
	Lang     string
	Class    string
}

type HtgoConfig struct {
	Router  *gin.Engine
	EmbedFS *embed.FS
	Options SetupOptions
}
