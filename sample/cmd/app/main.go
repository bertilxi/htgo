package main

import (
	"github.com/bertilxi/htgo"
	app "github.com/bertilxi/htgo/sample"
)

func main() {
	r := app.NewRouter()

	htgo.New(app.NewHtgoConfig(r))

	r.Run()
}
