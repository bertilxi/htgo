package app

import (
	"embed"
	"net/http"

	"github.com/bertilxi/htgo"
	"github.com/bertilxi/htgo/examples/sink/app/pages"
	"github.com/bertilxi/htgo/examples/sink/app/public"
	"github.com/gin-gonic/gin"
)

//go:embed .htgo
var EmbedFS embed.FS

func NewOptions(r *gin.Engine) htgo.Options {
	return htgo.Options{
		Router:   r,
		EmbedFS:  &EmbedFS,
		PagesDir: "./app/pages",
		Title:    "Picsel",
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
		Loaders: map[string]func(c *gin.Context) (any, error){
			"/":      pages.LoadIndex,
			"/about": pages.LoadAbout,
		},
	}
}

func NewEngine() *gin.Engine {
	r := gin.Default()

	r.StaticFS("/public", http.FS(public.Public))

	return r
}
