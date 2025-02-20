package main

import (
	"github.com/bertilxi/htgo"
	"github.com/bertilxi/htgo/cli"
	app "github.com/bertilxi/htgo/examples/sink"
)

func main() {
	r := app.NewEngine()

	cli.Dev(htgo.New(app.NewOptions(r)))
}
