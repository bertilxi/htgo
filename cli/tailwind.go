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

	"github.com/bertilxi/htgo"
	"github.com/evanw/esbuild/pkg/api"
)

const tailwindPath = "./.htgo-cache/tailwindcss"

func download(url string, path string) error {
	out, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
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

func getTailwindPath() (string, error) {
	if _, err := os.Stat(tailwindPath); os.IsNotExist(err) {
		os.MkdirAll("./.htgo-cache", 0755)
		fmt.Println("downloading tailwind")
		err = download(tailwindUrl(), tailwindPath)
		if err != nil {
			return "", err
		}
		fmt.Println("tailwind downloaded")
	}

	return tailwindPath, nil
}

func runTailwind(inputFile string, outputFile string, minify bool) error {
	cmdPath, err := getTailwindPath()

	if err != nil {
		return err
	}

	args := []string{
		"-i", inputFile,
		"-o", outputFile,
	}

	if minify {
		args = append(args, "-m")
	}

	cmd := exec.Command(cmdPath, args...)

	_, err = cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

func newTailwindPlugin(shouldMinify bool) api.Plugin {
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

				cwd, _ := os.Getwd()
				tmpFilePath := filepath.Join(
					cwd,
					htgo.CacheDir,
					strings.ReplaceAll(strings.ReplaceAll(sourceFullPath, ".css", ""), cwd, "")+".tmp.css",
				)
				tmpFiles = append(tmpFiles, tmpFilePath)
				err = runTailwind(sourceFullPath, tmpFilePath, shouldMinify)

				return api.OnResolveResult{Path: tmpFilePath}, err
			})
		},
	}
}
