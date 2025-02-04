package app

import (
	"github.com/bertilxi/htgo"
)

var HtgoOptions = htgo.SetupOptions{
	Title:   "Picsel",
	Lang:    "en",
	Class:   "dark",
	Hydrate: true,
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
		},
		{
			Route: "/about",
			File:  "./app/pages/about.tsx",
		},
	},
}
