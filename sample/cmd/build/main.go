package main

import (
	"os"

	"github.com/bertilxi/htgo"

	app "github.com/bertilxi/htgo/sample"
)

func main() {
	os.Setenv("GIN_MODE", "release")
	htgo.Build(app.HtgoOptions)
}
