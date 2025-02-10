package cli

import (
	"log"
	"os"
	"path"

	"github.com/bertilxi/htgo"
	"github.com/bertilxi/htgo/ssr"
)

func Build(config htgo.HtgoConfig) {
	cleanCache()

	config.Options = htgo.MapOptions(config.Options)

	for _, page := range config.Options.Pages {
		page = htgo.GetPage(page, config.Options)

		htgo.BuildBackendCached(page.File)
		htgo.BuildClientCached(page.File)

		if config.Options.Mode.Name != htgo.ModeSSR {
			ssr.SsrBuild(page)
		}
	}
}

func cleanCache() {
	err := os.RemoveAll(htgo.CacheDir)
	if err != nil {
		log.Fatal("Could not remove cache directory:", err)
	}

	err = os.MkdirAll(htgo.CacheDir, 0755)
	if err != nil {
		log.Fatal("Could not create cache directory:", err)
	}

	err = os.WriteFile(path.Join(htgo.CacheDir, "keep"), []byte(""), 0644)
	if err != nil {
		log.Fatal("Could not write keep file:", err)
	}
}
