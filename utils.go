package htgo

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const cacheDir = ".htgo"

func isDev() bool {
	return os.Getenv("GIN_MODE") == "debug"
}

func pageCacheKey(page string, extension string) string {
	pageKey := strings.TrimSuffix(page, filepath.Ext(page))
	cacheKey := fmt.Sprintf("%s.%s", pageKey, extension)
	return path.Join(cacheDir, cacheKey)
}

func mkdirCache(page string) {
	if err := os.MkdirAll(path.Dir(pageCacheKey(page, "")), 0755); err != nil {
		log.Fatal("Could not create cache directory:", err)
	}
}

var EmbedFS *embed.FS

func SetEmbedFS(fs *embed.FS) {
	EmbedFS = fs
}

func readFile(name string) ([]byte, error) {
	if isDev() || EmbedFS == nil {
		return os.ReadFile(name)
	}

	return EmbedFS.ReadFile(name)
}

func assignPage(page Page, newPage Page) Page {
	if newPage.Title != "" {
		page.Title = newPage.Title
	}
	if newPage.Lang != "" {
		page.Lang = newPage.Lang
	}
	if newPage.Class != "" {
		page.Class = newPage.Class
	}
	if newPage.MetaTags != nil {
		page.MetaTags = append(page.MetaTags, newPage.MetaTags...)
	}
	if newPage.Links != nil {
		page.Links = append(page.Links, newPage.Links...)
	}
	if newPage.Props != nil {
		page.Props = newPage.Props
	}

	return page
}

func getPage(page Page, options SetupOptions) Page {
	page.Lang = options.Lang
	page.Class = options.Class
	page.Links = append(page.Links, options.Links...)
	page.MetaTags = append(page.MetaTags, options.MetaTags...)

	if page.Title == "" {
		page.Title = options.Title
	}

	return page
}
