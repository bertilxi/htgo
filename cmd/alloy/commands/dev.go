package commands

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bertilxi/alloy"
	"github.com/bertilxi/alloy/cli"
)

func DevCmd(args []string) {
	fs := flag.NewFlagSet("dev", flag.ExitOnError)
	port := fs.String("port", "8080", "Port for dev server")
	dir := fs.String("dir", ".", "Project directory")

	fs.Parse(args)

	if err := runDev(*port, *dir); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Dev server error: %v\n", err)
		os.Exit(1)
	}
}

func runDev(port, dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	mainFilePath := filepath.Join(absDir, "main.go")
	if _, err := os.Stat(mainFilePath); err != nil {
		return fmt.Errorf("main.go not found in %s - are you in an Alloy project?", dir)
	}

	fmt.Printf("📁 Loading project from: %s\n", absDir)

	engine, err := loadEngine(absDir)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	if port != "8080" {
		engine.Port = port
	}

	fmt.Printf("🚀 Starting dev server on port %s...\n\n", port)

	return cli.Dev(engine)
}

func loadEngine(dir string) (*alloy.Engine, error) {
	// Read go.mod to get module name
	modPath := filepath.Join(dir, "go.mod")
	modContent, err := os.ReadFile(modPath)
	if err != nil {
		return nil, fmt.Errorf("go.mod not found in %s", dir)
	}

	moduleName := parseModuleName(string(modContent))
	if moduleName == "" {
		return nil, fmt.Errorf("could not parse module name from go.mod")
	}

	// Generate a temporary dev program
	devProgram := generateDevProgram(moduleName)

	// Create a temporary directory for the dev program
	tempDir, err := ioutil.TempDir("", "alloy-dev-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write the temporary dev program
	tempFile := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(tempFile, []byte(devProgram), 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp dev program: %w", err)
	}

	// Change to project directory to run go run from there
	originalDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(dir); err != nil {
		return nil, fmt.Errorf("failed to change directory: %w", err)
	}
	defer os.Chdir(originalDir)

	// Run the temporary program (use -mod=mod to allow running from temp directory)
	cmd := exec.Command("go", "run", "-mod=mod", tempFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// This will block until the dev server is interrupted
	if err := cmd.Run(); err != nil {
		// Don't return error on interrupt (user pressed Ctrl+C)
		if _, ok := err.(*exec.ExitError); ok {
			return nil, nil
		}
		return nil, err
	}

	return nil, nil
}

func parseModuleName(modContent string) string {
	for _, line := range strings.Split(modContent, "\n") {
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}

func generateDevProgram(moduleName string) string {
	return fmt.Sprintf(`package main

import (
	"github.com/bertilxi/alloy"
	"github.com/bertilxi/alloy/cli"
	"%s/pages"
)

func main() {
	options := alloy.Options{
		EmbedFS:  nil,
		Title:    "My Alloy App",
		Loaders:  pages.LoaderRegistry,
		Handlers: pages.HandlerRegistry,
	}
	if err := cli.Dev(alloy.New(options)); err != nil {
		panic(err)
	}
}
`, moduleName)
}
