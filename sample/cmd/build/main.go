package main

import (
	"github.com/bertilxi/htgo/cli"

	app "github.com/bertilxi/htgo/sample"
)

func main() {
	cli.Build(app.NewHtgoConfig(nil))
}
