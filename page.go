package htgo

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"

	"github.com/buke/quickjs-go"
	"github.com/gin-gonic/gin"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="{{.Lang}}" class="{{.Class}}">
<head>
    <meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{.Title}}</title>
	<link rel="stylesheet" href="{{.CSS}}" />
	{{range .MetaTags}}
		<meta name="{{.Name}}" content="{{.Content}}" property="{{.Property}}" />
	{{end}}
	{{range .Links}}
		<link rel="{{.Rel}}" href="{{.Href}}" />
	{{end}}
</head>
<body>
    <div id="page">{{.RenderedContent}}</div>
	{{if .Hydrate}}
	<script type="module" src="{{.JS}}"></script>
	<script>window.PAGE_PROPS = {{.InitialProps}};</script>
	{{end}}

	{{if .IsDev}}
	<script>
      function debounce(func, timeout = 500) {
        let timer;
        return (...args) => {
          clearTimeout(timer);
          timer = setTimeout(() => {
            func.apply(this, args);
          }, timeout);
        };
      }
      
      const reload = debounce(() => {
        console.log("reloading...");
        window.location.reload(true);
      });
      
      function start() {
        let socket = new WebSocket("ws://127.0.0.1:8080/ws");
      
        socket.onmessage = reload
      
        socket.onclose = () => {
          socket = null;
          setTimeout(start, 1000);
        };
      }
      
      start();
	</script>
	{{end}}
</body>
</html>`

func (page *Page) assignPage(newPage Page) {
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
}

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
}

func (page *Page) clone() Page {
	return Page{
		Route:       page.Route,
		File:        page.File,
		Interactive: page.Interactive,
		Props:       page.Props,
		Title:       page.Title,
		MetaTags:    page.MetaTags,
		Links:       page.Links,
		Lang:        page.Lang,
		Class:       page.Class,
		Handler:     page.Handler,
		embedFS:     page.embedFS,
	}
}

func (page *Page) readFile(name string) ([]byte, error) {
	if IsDev() || page.embedFS == nil {
		return os.ReadFile(name)
	}

	return page.embedFS.ReadFile(name)
}

func (page *Page) getServerJsFromFs() (string, error) {
	cacheKey := PageCacheKey(page.File, "ssr.js")

	cached, err := page.readFile(cacheKey)

	if err != nil {
		return "", err
	}

	return string(cached), nil
}

func (page *Page) getClientJsFromFs() (string, string, error) {
	jsCacheKey := PageCacheKey(page.File, "js")
	cssCacheKey := PageCacheKey(page.File, "css")

	_, jsErr := page.readFile(jsCacheKey)
	_, cssErr := page.readFile(cssCacheKey)
	if jsErr != nil || cssErr != nil {
		return "", "", jsErr
	}

	return jsCacheKey, cssCacheKey, nil
}

func (page *Page) ssr(props string) (string, error) {
	bundle, err := page.getServerJsFromFs()
	if err != nil {
		return "", err
	}

	rt := quickjs.NewRuntime()
	defer rt.Close()
	ctx := rt.NewContext()
	defer ctx.Close()

	res, err := ctx.Eval(bundle + "; renderPage(" + props + ")")
	if err != nil {
		return "", err
	}

	return res.String(), nil
}

func (p *Page) render(c *gin.Context) {
	page := p.clone()

	if page.Handler != nil {
		newPage := page.Handler(c)
		page.assignPage(newPage)
	}

	jsonProps, err := json.Marshal(page.Props)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	renderedHTML, err := page.ssr(string(jsonProps))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	clientBundle, clientCSS, err := page.getClientJsFromFs()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tmpl, err := template.New("webpage").Parse(htmlTemplate)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	data := htmlTemplateData{
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
		Hydrate:         page.Interactive,
	}

	c.Header("Content-Type", "text/html")

	err = tmpl.Execute(c.Writer, data)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}
