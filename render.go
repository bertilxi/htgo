package htgo

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/buke/quickjs-go"
	"github.com/gin-gonic/gin"
)

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
        const wsPort = "{{.WebSocketPort}}" || window.location.port || "8080";
        const wsUrl = "ws://" + window.location.hostname + ":" + wsPort + "/ws";
        let socket = new WebSocket(wsUrl);

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

func getErrorHandler(page *Page) ErrorHandler {
	if ctx, exists := pageContexts[page.File]; exists && ctx.errorHandler != nil {
		return ctx.errorHandler
	}
	return nil
}

func (p *Page) render(c *gin.Context) {
	errorHandler := getErrorHandler(p)
	props := p.Props

	if p.Handler != nil {
		handlerProps, err := p.Handler(c)
		if err != nil {
			if errorHandler != nil {
				errorHandler(c, err, p)
			} else {
				renderErr := &renderError{
					step:    "handler execution",
					message: "Handler failed",
					details: err.Error(),
				}
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": renderErr.Error(),
					"page":  p.Route,
				})
			}
			return
		}
		if handlerProps != nil {
			props = handlerProps
		}
	}

	jsonProps, err := json.Marshal(props)
	if err != nil {
		renderErr := &renderError{
			step:    "props serialization",
			message: "Failed to convert props to JSON",
			details: err.Error(),
		}
		if errorHandler != nil {
			errorHandler(c, renderErr, p)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": renderErr.Error(),
				"page":  p.Route,
			})
		}
		return
	}

	renderedHTML, err := p.ssr(string(jsonProps))
	if err != nil {
		details := extractJSErrorContext(err.Error())
		renderErr := &renderError{
			step:    "server-side rendering",
			message: "React component rendering failed",
			details: details,
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": renderErr.Error(),
			"page":  p.Route,
			"file":  p.File,
		})
		return
	}

	clientBundle, clientCSS, err := p.getClientJsFromFs()
	if err != nil {
		renderErr := &renderError{
			step:    "bundle loading",
			message: "Client bundle files not found",
			details: fmt.Sprintf("Expected files for: %s", p.File),
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": renderErr.Error(),
			"page":  p.Route,
			"file":  p.File,
		})
		return
	}

	tmpl, err := template.New("webpage").Parse(htmlTemplate)
	if err != nil {
		renderErr := &renderError{
			step:    "template parsing",
			message: "Internal template error",
			details: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": renderErr.Error(),
		})
		return
	}

	data := htmlTemplateData{
		RenderedContent: template.HTML(renderedHTML),
		InitialProps:    template.JS(jsonProps),
		JS:              template.JS(p.assetURL(clientBundle)),
		CSS:             template.CSS(p.assetURL(clientCSS)),
		Title:           template.HTML(p.Title),
		IsDev:           IsDev(),
		RouteID:         p.File,
		MetaTags:        p.MetaTags,
		Links:           p.Links,
		Lang:            template.HTML(p.Lang),
		Class:           template.HTML(p.Class),
		Hydrate:         p.Interactive,
		WebSocketPort:   p.port,
	}

	c.Header("Content-Type", "text/html")

	err = tmpl.Execute(c.Writer, data)
	if err != nil {
		renderErr := &renderError{
			step:    "template execution",
			message: "Failed to render HTML",
			details: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": renderErr.Error(),
		})
		return
	}
}
