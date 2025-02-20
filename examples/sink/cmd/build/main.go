package main

import (
	"github.com/bertilxi/htgo"
	"github.com/bertilxi/htgo/cli"
	app "github.com/bertilxi/htgo/examples/sink"
)

func main() {
	cli.Build(htgo.New(app.NewOptions(nil)))
}
