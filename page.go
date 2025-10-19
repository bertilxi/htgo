package alloy

import (
	"embed"
	"strings"
)

type pageContext struct {
	embedFS      *embed.FS
	errorHandler ErrorHandler
}

var pageContexts = make(map[string]pageContext)

func (page *Page) AssignOptions(options Options) {
	page.embedFS = options.EmbedFS
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

	pageContexts[page.File] = pageContext{
		embedFS:      options.EmbedFS,
		errorHandler: options.ErrorHandler,
	}
}

func (page *Page) assetURL(path string) string {
	url := "/" + path
	if strings.HasPrefix(url, "//") {
		url = strings.TrimPrefix(url, "/")
	}
	return url
}
