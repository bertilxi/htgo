package htgo

import (
	"embed"
	"strings"
)

type pageContext struct {
	embedFS          *embed.FS
	errorHandler     ErrorHandler
	assetURLPrefix   string
	cacheBustVersion string
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
		embedFS:          options.EmbedFS,
		errorHandler:     options.ErrorHandler,
		assetURLPrefix:   options.AssetURLPrefix,
		cacheBustVersion: options.CacheBustVersion,
	}
}

func (page *Page) assetURL(path string) string {
	ctx, exists := pageContexts[page.File]
	if !exists {
		ctx = pageContext{}
	}

	prefix := ctx.assetURLPrefix
	if prefix == "" {
		prefix = "/"
	}

	url := prefix + path
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if ctx.cacheBustVersion != "" {
		url += "?v=" + ctx.cacheBustVersion
	}
	return url
}
