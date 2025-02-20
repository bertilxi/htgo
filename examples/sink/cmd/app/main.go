package main

import (
	"github.com/bertilxi/htgo"
	app "github.com/bertilxi/htgo/examples/sink"
)

func main() {
	r := app.NewEngine()

	htgo.New(app.NewOptions(r)).Start()
}
