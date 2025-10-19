package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

type goWatcher struct {
	devBinary   string
	debounce    time.Duration
	lastRebuild time.Time
	watchDirs   []string
	sigChan     chan os.Signal
}

func newGoWatcher(devBinary string, sigChan chan os.Signal) *goWatcher {
	return &goWatcher{
		devBinary:   devBinary,
		debounce:    100 * time.Millisecond,
		lastRebuild: time.Now().Add(-1 * time.Second),
		watchDirs:   []string{".", "cmd", "app", "pages"},
		sigChan:     sigChan,
	}
}

func (gw *goWatcher) shouldRebuild() bool {
	return time.Since(gw.lastRebuild) > gw.debounce
}

func (gw *goWatcher) rebuild() error {
	if !gw.shouldRebuild() {
		return nil
	}

	gw.lastRebuild = time.Now()

	fmt.Println("🔄 Rebuilding...")

	cmd := exec.Command("go", "build", "-o", "tmp/bin/dev", gw.devBinary)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("❌ Build failed: %v\n", err)
		return err
	}

	fmt.Println("✓ Build successful, restarting server...")
	return nil
}

func (gw *goWatcher) restart() error {
	// Use syscall.Exec to replace current process with new binary
	// This is clean and avoids zombie processes on Unix systems
	binary, err := filepath.Abs("tmp/bin/dev")
	if err != nil {
		return err
	}

	// On success, syscall.Exec does not return
	err = syscall.Exec(binary, []string{binary}, os.Environ())
	if err != nil {
		fmt.Printf("❌ Restart failed: %v\n", err)
		return err
	}

	return nil
}

func (gw *goWatcher) isGoFile(path string) bool {
	if !strings.HasSuffix(path, ".go") {
		return false
	}
	// Skip build artifacts and generated files
	name := filepath.Base(path)
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, "_test.go") {
		return false
	}
	return true
}

func (gw *goWatcher) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Watch directories that exist
	for _, dir := range gw.watchDirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue // Skip non-existent or non-directory paths
		}

		// Add the directory itself
		watcher.Add(dir)

		// For root directory, only walk if it's not "."
		if dir == "." {
			continue
		}

		err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip paths we can't access
			}
			if fi.IsDir() && !strings.HasPrefix(fi.Name(), ".") {
				watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Warning: Could not watch directory %s: %v\n", dir, err)
		}
	}

	rebuildChan := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if gw.isGoFile(event.Name) && event.Op&fsnotify.Write == fsnotify.Write {
					select {
					case rebuildChan <- struct{}{}:
					default:
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				if err != nil {
					fmt.Printf("Watcher error: %v\n", err)
				}
			}
		}
	}()

	for {
		select {
		case <-rebuildChan:
			err := gw.rebuild()
			if err == nil {
				err = gw.restart()
				if err != nil {
					// On Windows or other systems where Exec doesn't work,
					// this will fail. For now, we just log it.
					fmt.Printf("Note: Process restart not supported on this platform\n")
				}
			}

		case <-gw.sigChan:
			fmt.Println("\nShutdown signal received")
			return nil
		}
	}
}
