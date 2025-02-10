package cli

import (
	"github.com/bertilxi/htgo"

	"github.com/bertilxi/htgo/ssr"
)

func Dev(config htgo.HtgoConfig) {
	appPages := htgo.GetPages(config.Options)

	for _, page := range appPages {
		htgo.MkdirCache(page.File)

		go htgo.WatchServer(page)
		go htgo.WatchClient(page)
		go htgo.StartWatcher()
	}

	config.Router.Static(htgo.CacheDir, htgo.CacheDir)
	config.Router.GET("/ws", htgo.WebsocketHandler)
	config.Options.Mode = &ssr.HtgoModeSSR

	htgo.New(config)
}
