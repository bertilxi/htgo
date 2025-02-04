package htgo

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

func RenderPage(page Page) func(c *gin.Context) {
	return func(c *gin.Context) {
		if page.Handler != nil {
			newPage := page.Handler(c)
			page = assignPage(page, newPage)
		}

		jsonProps, err := json.Marshal(page.Props)

		if err != nil {
			log.Fatal("Error parsing props:", err)
		}

		renderedHTML := ssr(page.File, string(jsonProps))
		clientBundle, clientCSS := buildClientCached(page.File)

		tmpl, err := template.New("webpage").Parse(htmlTemplate)
		if err != nil {
			log.Fatal("Error parsing template:", err)
		}

		data := htmlTemplateData{
			RenderedContent: template.HTML(renderedHTML),
			InitialProps:    template.JS(jsonProps),
			JS:              template.JS(clientBundle),
			CSS:             template.CSS(clientCSS),
			Title:           template.HTML(page.Title),
			IsDev:           isDev(),
			RouteID:         page.File,
			MetaTags:        page.MetaTags,
			Links:           page.Links,
			Lang:            template.HTML(page.Lang),
			Class:           template.HTML(page.Class),
		}

		c.Header("Content-Type", "text/html")
		err = tmpl.Execute(c.Writer, data)

		if err != nil {
			log.Fatal("Error executing template:", err)
		}
	}
}

func New(config HtgoConfig) {
	router := config.Router
	options := config.Options
	EmbedFS := config.EmbedFS

	if EmbedFS != nil {
		SetEmbedFS(EmbedFS)
	}

	appPages := []Page{}

	for _, page := range options.Pages {
		appPages = append(appPages, getPage(page, options))
	}

	for _, page := range appPages {
		router.GET(page.Route, RenderPage(page))
	}

	if !isDev() {
		router.Any(cacheDir+"/*path", func(c *gin.Context) {
			route := c.Param("path")

			c.FileFromFS(path.Join(cacheDir, route), http.FS(EmbedFS))
		})
	}

	if isDev() {
		router.Static(cacheDir, cacheDir)
		router.GET("/ws", websocketHandler)

		for _, page := range appPages {
			mkdirCache(page.File)

			go watchServer(page)
			go watchClient(page)
			go startWatcher()
		}
	}
}

func Build(options SetupOptions) {
	for _, page := range options.Pages {
		page = getPage(page, options)

		buildBackendCached(page.File)
		buildClientCached(page.File)
	}
}
