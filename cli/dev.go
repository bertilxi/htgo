package cli

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/bertilxi/alloy"
	"github.com/bertilxi/alloy/core"
)

func mkdirCache(page string) error {
	err := os.MkdirAll(path.Dir(core.PageCacheKey(page, "")), 0755)
	if err != nil {
		return err
	}

	return nil
}

func Dev(engine *alloy.Engine) error {
	err := core.CleanCache()
	if err != nil {
		return err
	}

	// Generate loader registry from .go files
	err = ensureGeneratedLoaders(engine.Options.PagesDir)
	if err != nil {
		return err
	}

	// Discover pages first
	pages, err := alloy.DiscoverPages(engine.Options.PagesDir, engine.Options.Loaders)
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

	// Register bundles and routes
	engine.RegisterBundles()
	if err := engine.RegisterRoutes(); err != nil {
		return err
	}
	engine.Router.GET("/ws", hr.websocket)

	// Print dev server ready message with routes
	port := engine.Port
	if port == "" {
		port = "8080"
	}
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✓ Alloy Dev Server Ready")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("🌐 Local:       http://localhost:%s\n", port)
	fmt.Println()
	fmt.Println("📄 Routes:")
	for _, page := range engine.Pages {
		fmt.Printf("   • %s\n", page.Route)
	}
	fmt.Println()
	fmt.Println("🔄 Hot reload enabled - changes will auto-refresh")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Start the HTTP server in a goroutine
	go func() {
		engine.Listen()
	}()

	// Start watching Go files (blocks until signal received)
	return gw.watch()
}
