package main

import (
	"github.com/bertilxi/htgo/cli"

	app "github.com/bertilxi/htgo/sample"
)

func main() {
	r := app.NewRouter()

	cli.Dev(app.NewHtgoConfig(r))

	r.Run()
}
