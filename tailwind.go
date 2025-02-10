package htgo

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

func download(url string, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	err = os.Chmod(path, 0755)
	if err != nil {
		return err
	}

	return nil
}

func tailwindUrl() string {
	goos := runtime.GOOS
	arch := runtime.GOARCH

	if goos == "windows" && arch == "amd64" {
		return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-windows-x64.exe"
	}
	if goos == "linux" && arch == "arm64" {
		return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-arm64"
	}
	if goos == "linux" && arch == "amd64" {
		return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64"
	}
	if goos == "darwin" && arch == "arm64" {
		return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64"
	}
	if goos == "darwin" && arch == "amd64" {
		return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64"
	}

	return "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64"
}

func getTailwindPath() string {
	if _, err := os.Stat("./.htgo-cache/tailwindcss"); os.IsNotExist(err) {
		os.MkdirAll("./.htgo-cache", 0755)

		fmt.Println("downloading tailwind")
		download(tailwindUrl(), "./.htgo-cache/tailwindcss")
		fmt.Println("tailwind downloaded")
	}

	return "./.htgo-cache/tailwindcss"
}

func runTailwind(inputFile string, outputFile string, minify bool) error {
	cmdPath := getTailwindPath()

	args := []string{
		"-i", inputFile,
		"-o", outputFile,
	}

	if minify {
		args = append(args, "-m")
	}

	cmd := exec.Command(cmdPath, args...)

	_, err := cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

func NewTailwindPlugin(shouldMinify bool) api.Plugin {
	go getTailwindPath()

	return api.Plugin{
		Name: "tailwind",
		Setup: func(b api.PluginBuild) {
			tmpFiles := []string{}

			b.OnResolve(api.OnResolveOptions{
				Filter:    `.\.(css)$`,
				Namespace: "file",
			}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
				sourceFullPath := filepath.Join(args.ResolveDir, args.Path)
				source, err := os.ReadFile(sourceFullPath)
				if err != nil {
					return api.OnResolveResult{Path: sourceFullPath}, err
				}

				if !strings.Contains(string(source), `@import "tailwindcss"`) {
					return api.OnResolveResult{Path: sourceFullPath}, nil
				}

				tmpFile := strings.ReplaceAll(
					filepath.Base(sourceFullPath),
					filepath.Ext(sourceFullPath),
					"") + ".tmp.css"
				tmpFilePath := filepath.Join(filepath.Dir(sourceFullPath), tmpFile)
				tmpFiles = append(tmpFiles, tmpFilePath)

				err = runTailwind(sourceFullPath, tmpFilePath, shouldMinify)
				return api.OnResolveResult{
					Path: tmpFilePath,
				}, err
			})
			b.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
				for _, tmp := range tmpFiles {
					_ = os.Remove(tmp)
				}

				return api.OnEndResult{
					Errors:   result.Errors,
					Warnings: result.Warnings,
				}, nil
			})
		},
	}
}
