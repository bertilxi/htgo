package app

import (
	"embed"
	"net/http"
	"time"

	"github.com/bertilxi/htgo"
	"github.com/bertilxi/htgo/examples/sink/app/public"
	"github.com/gin-gonic/gin"
)

//go:embed .htgo
var EmbedFS embed.FS

func NewOptions(r *gin.Engine) htgo.Options {
	return htgo.Options{
		Router:  r,
		EmbedFS: &EmbedFS,
		Title:   "Picsel",
		MetaTags: []htgo.MetaTag{
			{
				Name:     "description",
				Content:  "Picsel is a simple image selector",
				Property: "og:description",
			},
			{
				Name:     "keywords",
				Content:  "image selector, image picker, image gallery",
				Property: "og:keywords",
			},
			{
				Name:     "og:title",
				Content:  "Picsel",
				Property: "og:title",
			},
		},
		Links: []htgo.Link{
			{
				Rel:  "icon",
				Href: "/public/favicon.ico",
			},
		},
		Pages: []htgo.Page{
			{
				Route:       "/",
				File:        "./app/pages/index.tsx",
				Interactive: true,
				Handler: func(c *gin.Context) htgo.Page {
					return htgo.Page{
						Props: map[string]any{
							"route": c.FullPath(),
							"time":  time.Now().String(),
						},
					}
				},
			},
			{
				Route: "/about",
				File:  "./app/pages/about.tsx",
			},
		},
	}
}

func NewEngine() *gin.Engine {
	r := gin.Default()

	r.StaticFS("/public", http.FS(public.Public))

	return r
}
