package main

import (
	"os"

	"github.com/bertilxi/htgo"
	app "github.com/bertilxi/htgo/sample"
	"github.com/gin-gonic/gin"
)

func main() {
	os.Setenv("GIN_MODE", "release")
	// os.Setenv("GIN_MODE", "debug")

	r := gin.Default()

	htgo.New(htgo.HtgoConfig{
		Router:  r,
		Options: app.HtgoOptions,
		EmbedFS: &app.EmbedFS,
	})

	r.Static("/public", "./app/public")

	r.Run()
}
