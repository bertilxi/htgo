package app

import (
	"embed"

	"github.com/bertilxi/htgo"
)

//go:embed .htgo
var EmbedFS embed.FS

var Options = htgo.Options{
	EmbedFS:  &EmbedFS,
	PagesDir: "./pages",
}
