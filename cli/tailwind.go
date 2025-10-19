package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/bertilxi/alloy"
	"github.com/bertilxi/alloy/core"
	"github.com/evanw/esbuild/pkg/api"
)

const tailwindPath = "./.alloy-cache/tailwindcss"

// Tailwind CSS support:
// - Automatically downloads the Tailwind CLI on first build for your platform
// - Only CSS files with @import "tailwindcss" are processed (others pass through unchanged)
// - Server bundles exclude CSS entirely (LoaderEmpty)
// - Client bundles include processed CSS as separate .css file
// - In development, changes to CSS rebuild automatically
// - In production, CSS output is minified
//
// Safe by design:
// - CSS files without Tailwind directive are never modified
// - Graceful error handling with clear messages if Tailwind processing fails
// - Temporary files are isolated in .alloy-cache directory
// - No configuration required - uses Tailwind v4 defaults with inline @theme/@plugin support

func download(url string, path string) error {
	out, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create tailwind binary file: %w", err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download tailwind from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tailwind download failed with status %d", resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write tailwind binary: %w", err)
	}

	return nil
}

func tailwindUrl() (string, error) {
	goos := runtime.GOOS
	arch := runtime.GOARCH

	switch {
	case goos == "windows" && arch == "amd64":
		return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-windows-x64.exe", nil
	case goos == "linux" && arch == "arm64":
		return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-arm64", nil
	case goos == "linux" && arch == "amd64":
		return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64", nil
	case goos == "darwin" && arch == "arm64":
		return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64", nil
	case goos == "darwin" && arch == "amd64":
		return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, arch)
	}
}

func getTailwindPath() (string, error) {
	if _, err := os.Stat(tailwindPath); os.IsNotExist(err) {
		if err := os.MkdirAll("./.alloy-cache", 0755); err != nil {
			return "", fmt.Errorf("failed to create cache directory: %w", err)
		}

		url, err := tailwindUrl()
		if err != nil {
			return "", err
		}

		fmt.Println("downloading tailwind...")
		err = download(url, tailwindPath)
		if err != nil {
			return "", err
		}
		fmt.Println("tailwind downloaded successfully")
	}

	return tailwindPath, nil
}

func runTailwind(inputFile string, outputFile string, minify bool) error {
	cmdPath, err := getTailwindPath()
	if err != nil {
		return fmt.Errorf("failed to get tailwind path: %w", err)
	}

	args := []string{
		"-i", inputFile,
		"-o", outputFile,
	}

	if minify {
		args = append(args, "-m")
	}

	cmd := exec.Command(cmdPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tailwind processing failed: %w\noutput: %s", err, string(output))
	}

	return nil
}

// DetectTailwind scans all pages to see if any CSS file uses Tailwind directive.
// Returns true if Tailwind CSS is used anywhere in the project.
func DetectTailwind(pages []alloy.Page) (bool, error) {
	// Check root-level CSS files
	if content, err := os.ReadFile("styles.css"); err == nil {
		if strings.Contains(string(content), `@import "tailwindcss"`) {
			return true, nil
		}
	}

	// Check CSS files in page directories
	for _, page := range pages {
		pageDir := filepath.Dir(page.File)
		entries, err := os.ReadDir(pageDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".css") {
				continue
			}

			cssPath := filepath.Join(pageDir, entry.Name())
			content, err := os.ReadFile(cssPath)
			if err != nil {
				continue
			}

			if strings.Contains(string(content), `@import "tailwindcss"`) {
				return true, nil
			}
		}
	}

	return false, nil
}

// EnsureTailwind detects if Tailwind is used and pre-downloads the CLI if needed.
// This is called before serving any pages to avoid lazy downloads during requests.
func EnsureTailwind(pages []alloy.Page) error {
	usesTailwind, err := DetectTailwind(pages)
	if err != nil {
		return fmt.Errorf("failed to detect tailwind usage: %w", err)
	}

	if !usesTailwind {
		return nil
	}

	_, err = getTailwindPath()
	if err != nil {
		return fmt.Errorf("failed to ensure tailwind is available: %w", err)
	}

	return nil
}

func newTailwindPlugin(shouldMinify bool, enableCache bool) api.Plugin {
	type cacheEntry struct {
		outputPath string
		modTime    int64
	}
	processingCache := make(map[string]cacheEntry)
	cacheMutex := sync.Mutex{}

	return api.Plugin{
		Name: "tailwind",
		Setup: func(b api.PluginBuild) {
			b.OnResolve(api.OnResolveOptions{
				Filter:    `.\.(css)$`,
				Namespace: "file",
			}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
				sourceFullPath := filepath.Join(args.ResolveDir, args.Path)
				source, err := os.ReadFile(sourceFullPath)
				if err != nil {
					return api.OnResolveResult{Path: sourceFullPath}, fmt.Errorf("failed to read CSS file %s: %w", sourceFullPath, err)
				}

				hasTailwind := strings.Contains(string(source), `@import "tailwindcss"`)
				if !hasTailwind {
					return api.OnResolveResult{Path: sourceFullPath}, nil
				}

				// Check file modification time
				info, err := os.Stat(sourceFullPath)
				if err != nil {
					return api.OnResolveResult{Path: sourceFullPath}, fmt.Errorf("failed to stat CSS file %s: %w", sourceFullPath, err)
				}
				currentModTime := info.ModTime().Unix()

				cacheMutex.Lock()
				cachedEntry, exists := processingCache[sourceFullPath]
				cacheMutex.Unlock()

				// Use cache only if enabled and file hasn't been modified
				if enableCache && exists && cachedEntry.modTime == currentModTime {
					return api.OnResolveResult{Path: cachedEntry.outputPath}, nil
				}

				cwd, err := os.Getwd()
				if err != nil {
					return api.OnResolveResult{Path: sourceFullPath}, fmt.Errorf("failed to get working directory: %w", err)
				}

				tmpFilePath := filepath.Join(
					cwd,
					core.CacheDir,
					strings.ReplaceAll(strings.ReplaceAll(sourceFullPath, ".css", ""), cwd, "")+".tmp.css",
				)

				err = runTailwind(sourceFullPath, tmpFilePath, shouldMinify)
				if err != nil {
					return api.OnResolveResult{Path: sourceFullPath}, fmt.Errorf("tailwind plugin error for %s: %w", sourceFullPath, err)
				}

				cacheMutex.Lock()
				processingCache[sourceFullPath] = cacheEntry{
					outputPath: tmpFilePath,
					modTime:    currentModTime,
				}
				cacheMutex.Unlock()

				return api.OnResolveResult{Path: tmpFilePath}, nil
			})
		},
	}
}
