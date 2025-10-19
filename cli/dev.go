package cli

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

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

	// Generate loader registry from .go files
	err = ensureGeneratedLoaders(engine.Options.PagesDir)
	if err != nil {
		return err
	}

	// Discover pages first
	pages, err := htgo.DiscoverPages(engine.Options.PagesDir, engine.Options.Loaders)
	if err != nil {
		return err
	}
	engine.Pages = pages

	// Ensure Tailwind is available before starting dev server
	err = EnsureTailwind(engine.Pages)
	if err != nil {
		return err
	}

	// Create cache directories and do initial builds for all pages
	for _, page := range engine.Pages {
		err := mkdirCache(page.File)
		if err != nil {
			return err
		}

		b := bundler{page: &page}

		// Do initial build before watching
		fmt.Printf("📦 Building %s...\n", page.File)
		_, err = b.buildBackend()
		if err != nil {
			return err
		}

		_, _, err = b.buildClient()
		if err != nil {
			return err
		}
		fmt.Printf("✓ Built bundles for %s\n", page.File)

		go b.watch()
	}

	hr := newHotReload()

	go hr.watch()

	// Watch pages directory for new/renamed/deleted files
	pw := newPagesWatcher(engine, hr)
	go pw.watch()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Watch Go files and rebuild on changes
	gw := newGoWatcher("cmd/dev/main.go", sigChan)

	engine.Router.Static(htgo.CacheDir, htgo.CacheDir)
	engine.Router.GET("/ws", hr.websocket)

	// Start the HTTP server in a goroutine
	go func() {
		engine.HandleRoutes()
	}()

	// Start watching Go files (blocks until signal received)
	return gw.watch()
}
