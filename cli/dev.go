package cli

import (
	"os"
	"path"

	"github.com/bertilxi/htgo"
)

func mkdirCache(page string) error {
	err := os.MkdirAll(path.Dir(htgo.PageCacheKey(page, "")), 0755)
	if err != nil {
		return err
	}

	return nil
}

func Dev(engine *htgo.Engine) error {
	err := htgo.CleanCache()
	if err != nil {
		return err
	}

	// Ensure Tailwind is available before starting dev server
	err = EnsureTailwind(engine.Pages)
	if err != nil {
		return err
	}

	for _, page := range engine.Pages {
		err := mkdirCache(page.File)
		if err != nil {
			return err
		}

		b := bundler{page: &page}

		go b.watch()
	}

	hr := newHotReload()

	go hr.watch()

	engine.Router.Static(htgo.CacheDir, htgo.CacheDir)
	engine.Router.GET("/ws", hr.websocket)
	engine.HandleRoutes()

	return nil
}
