package alloy

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
	<link rel="icon" href="/.alloy/favicon.svg" type="image/svg+xml" />
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
      let reconnectAttempts = 0;
      let reconnectDelay = 500;
      const maxReconnectDelay = 5000;

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

      let isFirstConnection = true;

      function start() {
        const wsPort = "{{.WebSocketPort}}" || window.location.port || "8080";
        const wsUrl = "ws://" + window.location.hostname + ":" + wsPort + "/ws";
        let socket = new WebSocket(wsUrl);

        socket.onopen = () => {
          // If reconnecting after a disconnect, reload the page
          if (reconnectAttempts > 0) {
            console.log("reconnected, reloading...");
            reload();
          }
          reconnectAttempts = 0;
          reconnectDelay = 500;
          isFirstConnection = false;
        };

        socket.onmessage = reload;

        socket.onerror = () => {
          socket.close();
        };

        socket.onclose = () => {
          socket = null;
          reconnectAttempts++;
          const delay = Math.min(reconnectDelay * Math.pow(1.5, reconnectAttempts), maxReconnectDelay);
          setTimeout(start, delay);
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

	res := ctx.Eval(bundle + "; renderPage(" + props + ")")
	defer res.Free()

	return res.String(), nil
}

func (p *Page) Render(c *gin.Context) {
	errorHandler := p.ErrorHandler
	props := p.Props

	if p.Loader != nil {
		loaderProps, err := p.Loader(c)
		if err != nil {
			if errorHandler != nil {
				errorHandler(c, err, p)
			} else {
				renderErr := &renderError{
					step:    "loader execution",
					message: "Loader failed",
					details: err.Error(),
				}
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": renderErr.Error(),
					"page":  p.Route,
				})
			}
			return
		}
		if loaderProps != nil {
			props = loaderProps
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
		WebSocketPort:   "", // Will use window.location.port or 8080
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
