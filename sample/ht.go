package app

import (
	"time"

	"github.com/bertilxi/htgo"
	"github.com/gin-gonic/gin"
)

var NewHtgoConfig = func(r *gin.Engine) htgo.HtgoConfig {
	return htgo.HtgoConfig{
		Router:  r,
		EmbedFS: &EmbedFS,
		Options: htgo.SetupOptions{
			Title: "Picsel",
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
					Route: "/",
					File:  "./app/pages/home.tsx",
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
		},
	}
}

func NewRouter() *gin.Engine {
	r := gin.Default()

	r.Static("/public", "./app/public")

	return r
}
