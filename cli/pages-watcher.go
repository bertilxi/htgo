package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bertilxi/htgo"
	"github.com/fsnotify/fsnotify"
)

type pagesWatcher struct {
	engine    *htgo.Engine
	pagesDir  string
	debounce  time.Duration
	lastEvent time.Time
	hotReload *hotReload
	mu        sync.Mutex
}

func newPagesWatcher(engine *htgo.Engine, hotReload *hotReload) *pagesWatcher {
	return &pagesWatcher{
		engine:    engine,
		pagesDir:  engine.PagesDir,
		debounce:  200 * time.Millisecond,
		lastEvent: time.Now(),
		hotReload: hotReload,
		mu:        sync.Mutex{},
	}
}

func (pw *pagesWatcher) shouldProcessEvent() bool {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	elapsed := time.Since(pw.lastEvent)
	if elapsed < pw.debounce {
		return false
	}

	pw.lastEvent = time.Now()
	return true
}

func (pw *pagesWatcher) isTsxFile(path string) bool {
	return strings.HasSuffix(path, ".tsx")
}

func (pw *pagesWatcher) processPageChanges() error {
	newPages, err := htgo.DiscoverPages(pw.pagesDir, pw.engine.Loaders)
	if err != nil {
		fmt.Printf("❌ Failed to discover pages: %v\n", err)
		return err
	}

	// Find added/modified pages
	newPageMap := make(map[string]*htgo.Page)
	for i := range newPages {
		newPageMap[newPages[i].Route] = &newPages[i]
	}

	oldPageMap := make(map[string]*htgo.Page)
	for i := range pw.engine.Pages {
		oldPageMap[pw.engine.Pages[i].Route] = &pw.engine.Pages[i]
	}

	// Check for new or modified pages
	for route, newPage := range newPageMap {
		if oldPage, exists := oldPageMap[route]; !exists {
			// New page
			fmt.Printf("📄 New page detected: %s (%s)\n", route, newPage.File)
			pw.registerNewPage(newPage)
		} else if oldPage.File != newPage.File {
			// Page file changed
			fmt.Printf("✏️  Page modified: %s\n", route)
		}
	}

	// Check for deleted pages
	for route := range oldPageMap {
		if _, exists := newPageMap[route]; !exists {
			fmt.Printf("🗑️  Page removed: %s\n", route)
		}
	}

	// Update engine pages
	pw.engine.Pages = newPages

	// Trigger hot reload
	pw.hotReload.reload()

	return nil
}

func (pw *pagesWatcher) registerNewPage(page *htgo.Page) error {
	// Assign engine options to the page
	page.AssignOptions(pw.engine.Options)

	// Create cache directory for the new page
	err := os.MkdirAll(filepath.Dir(htgo.PageCacheKey(page.File, "")), 0755)
	if err != nil {
		fmt.Printf("❌ Failed to create cache directory: %v\n", err)
		return err
	}

	// Register route with Gin
	pw.engine.Router.GET(page.Route, page.Render)

	// Create and start bundler for the new page
	b := bundler{page: page}

	// Do initial build
	fmt.Printf("📦 Building %s...\n", page.File)
	_, err = b.buildBackend()
	if err != nil {
		fmt.Printf("❌ Backend build failed: %v\n", err)
		return err
	}

	_, _, err = b.buildClient()
	if err != nil {
		fmt.Printf("❌ Client build failed: %v\n", err)
		return err
	}
	fmt.Printf("✓ Built bundles for %s\n", page.File)

	// Start watching the new page
	go b.watch()

	return nil
}

func (pw *pagesWatcher) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Watch the pages directory recursively
	err = filepath.Walk(pw.pagesDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() && !strings.HasPrefix(fi.Name(), ".") {
			return watcher.Add(path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Warning: Could not watch pages directory: %v\n", err)
	}

	// Also watch the root pages directory itself for new subdirectories
	watcher.Add(pw.pagesDir)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Detect when new directories are added to watch them
			if event.Op&fsnotify.Create == fsnotify.Create {
				fi, err := os.Stat(event.Name)
				if err == nil && fi.IsDir() && !strings.HasPrefix(fi.Name(), ".") {
					watcher.Add(event.Name)
					fmt.Printf("👀 Watching new directory: %s\n", event.Name)
				}
			}

			// Process .tsx file changes (create, remove, rename)
			if pw.isTsxFile(event.Name) {
				if event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					if pw.shouldProcessEvent() {
						fmt.Println("🔄 Pages directory changed, reprocessing...")
						err := pw.processPageChanges()
						if err != nil {
							fmt.Printf("⚠️  Error processing page changes: %v\n", err)
						}
					}
				}
			}

			// .go loader files are handled by the Go watcher, which triggers a rebuild
			// The rebuild will include the updated loader, and hot reload will happen automatically

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}
