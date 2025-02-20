package app

import (
	"github.com/bertilxi/htgo"
)

var Options = htgo.Options{
	Pages: []htgo.Page{
		{
			Route: "/",
			File:  "./pages/index.tsx",
		},
	},
}
