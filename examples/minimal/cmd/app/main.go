package main

import (
	"github.com/bertilxi/htgo"
	app "github.com/bertilxi/htgo/examples/minimal"
)

func main() {
	htgo.New(app.Options).Start()
}
