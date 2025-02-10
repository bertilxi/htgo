package ssr

import (
	"encoding/json"
	"html/template"
	"log"
	"os"

	"github.com/bertilxi/htgo"
	"github.com/gin-gonic/gin"
	v8 "rogchap.com/v8go"
)

var HtgoModeSSR = htgo.HtgoMode{
	Name:       htgo.ModeSSR,
	RenderPage: renderPage,
}

func ssr(page string, props string) string {
	backendBundle := htgo.BuildBackendCached(page)

	ctx := v8.NewContext()
	_, err := ctx.RunScript(backendBundle, "bundle.js")
	if err != nil {
		log.Fatal("Failed to evaluate bundled script:", err)
	}

	val, err := ctx.RunScript("renderPage("+props+")", "render.js")
	if err != nil {
		log.Fatal("Failed to render React component:", err)
	}

	return val.String()
}

func SsrBuild(page htgo.Page) string {
	if page.Handler != nil {
		newPage := page.Handler(&gin.Context{})
		page = htgo.AssignPage(page, newPage)
	}

	jsonProps, err := json.Marshal(page.Props)

	if err != nil {
		log.Fatal("Error parsing props:", err)
	}

	cacheKey := htgo.PageCacheKey(page.File, "html")

	result := ssr(page.File, string(jsonProps))

	if err := os.WriteFile(cacheKey, []byte(result), 0644); err != nil {
		log.Fatal("Could not write html to cache:", err)
	}

	return cacheKey
}

func renderPage(page htgo.Page) func(c *gin.Context) {
	return func(c *gin.Context) {
		if page.Handler != nil {
			newPage := page.Handler(c)
			page = htgo.AssignPage(page, newPage)
		}

		jsonProps, err := json.Marshal(page.Props)

		if err != nil {
			log.Fatal("Error parsing props:", err)
		}

		renderedHTML := ssr(page.File, string(jsonProps))
		clientBundle, clientCSS := htgo.BuildClientCached(page.File)

		tmpl, err := template.New("webpage").Parse(htgo.HtmlTemplate)
		if err != nil {
			log.Fatal("Error parsing template:", err)
		}

		data := htgo.HtmlTemplateData{
			RenderedContent: template.HTML(renderedHTML),
			InitialProps:    template.JS(jsonProps),
			JS:              template.JS(clientBundle),
			CSS:             template.CSS(clientCSS),
			Title:           template.HTML(page.Title),
			IsDev:           htgo.IsDev(),
			RouteID:         page.File,
			MetaTags:        page.MetaTags,
			Links:           page.Links,
			Lang:            template.HTML(page.Lang),
			Class:           template.HTML(page.Class),
			Hydrate:         page.Mode == htgo.PageModeJS,
		}

		c.Header("Content-Type", "text/html")
		err = tmpl.Execute(c.Writer, data)

		if err != nil {
			log.Fatal("Error executing template:", err)
		}
	}
}
