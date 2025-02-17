package htgo

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const CacheDir = ".htgo"

func IsDev() bool {
	return os.Getenv("GIN_MODE") != "release"
}

func PageCacheKey(page string, extension string) string {
	pageKey := strings.TrimSuffix(page, filepath.Ext(page))
	cacheKey := fmt.Sprintf("%s.%s", pageKey, extension)
	return path.Join(CacheDir, cacheKey)
}

func MkdirCache(page string) {
	if err := os.MkdirAll(path.Dir(PageCacheKey(page, "")), 0755); err != nil {
		log.Fatal("Could not create cache directory:", err)
	}
}

var embedFS *embed.FS

func SetEmbedFS(fs *embed.FS) {
	embedFS = fs
}

func readFile(name string) ([]byte, error) {
	if IsDev() || embedFS == nil {
		return os.ReadFile(name)
	}

	return embedFS.ReadFile(name)
}

func AssignPage(page Page, newPage Page) Page {
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

func GetPage(page Page, options SetupOptions) Page {
	page.Class = options.Class
	page.Links = append(page.Links, options.Links...)
	page.MetaTags = append(page.MetaTags, options.MetaTags...)
	page.Lang = options.Lang

	if page.Mode == "" {
		page.Mode = PageModeNoJS
	}
	if page.Lang == "" {
		page.Lang = "en"
	}
	if page.Title == "" {
		page.Title = options.Title
	}

	return page
}

func CopyPage(page Page) Page {
	return Page{
		Route:    page.Route,
		File:     page.File,
		Mode:     page.Mode,
		Props:    page.Props,
		Title:    page.Title,
		MetaTags: page.MetaTags,
		Links:    page.Links,
		Lang:     page.Lang,
		Class:    page.Class,
		Handler:  page.Handler,
	}
}

func renderPage(p Page) func(c *gin.Context) {
	return func(c *gin.Context) {
		page := CopyPage(p)

		if page.Handler != nil {
			newPage := page.Handler(c)
			page = AssignPage(page, newPage)
		}

		jsonProps, err := json.Marshal(page.Props)

		if err != nil {
			log.Fatal("Error parsing props:", err)
		}

		renderedHTML := ssrCached(page.File)
		clientBundle, clientCSS := BuildClientCached(page.File)

		tmpl, err := template.New("webpage").Parse(HtmlTemplate)
		if err != nil {
			log.Fatal("Error parsing template:", err)
		}

		data := HtmlTemplateData{
			RenderedContent: template.HTML(renderedHTML),
			InitialProps:    template.JS(jsonProps),
			JS:              template.JS(clientBundle),
			CSS:             template.CSS(clientCSS),
			Title:           template.HTML(page.Title),
			IsDev:           IsDev(),
			RouteID:         page.File,
			MetaTags:        page.MetaTags,
			Links:           page.Links,
			Lang:            template.HTML(page.Lang),
			Class:           template.HTML(page.Class),
			Hydrate:         page.Mode == PageModeJS,
		}

		c.Header("Content-Type", "text/html")
		err = tmpl.Execute(c.Writer, data)

		if err != nil {
			log.Fatal("Error executing template:", err)
		}
	}
}

func MapOptions(options SetupOptions) SetupOptions {
	if options.Mode == nil {
		options.Mode = &HtgoModeStatic
	}

	return options
}

func GetPages(options SetupOptions) []Page {
	appPages := []Page{}

	for _, page := range options.Pages {
		appPages = append(appPages, GetPage(page, options))
	}

	return appPages
}
