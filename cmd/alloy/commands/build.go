package commands

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func BuildCmd(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	dir := fs.String("dir", ".", "Project directory")
	output := fs.String("output", "", "Output binary path")

	fs.Parse(args)

	if err := runBuild(*dir, *output); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Build error: %v\n", err)
		os.Exit(1)
	}
}

func runBuild(dir, output string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	mainFilePath := filepath.Join(absDir, "main.go")
	if _, err := os.Stat(mainFilePath); err != nil {
		return fmt.Errorf("main.go not found in %s - are you in an Alloy project?", dir)
	}

	fmt.Printf("📁 Building project from: %s\n", absDir)

	// Generate temporary build program
	modPath := filepath.Join(absDir, "go.mod")
	modContent, err := os.ReadFile(modPath)
	if err != nil {
		return fmt.Errorf("go.mod not found in %s", dir)
	}

	moduleName := parseModuleName(string(modContent))
	if moduleName == "" {
		return fmt.Errorf("could not parse module name from go.mod")
	}

	buildProgram := generateBuildProgram(moduleName)

	// Create a temporary directory for the build program
	tempDir, err := ioutil.TempDir("", "alloy-build-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write the temporary build program
	tempFile := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(tempFile, []byte(buildProgram), 0644); err != nil {
		return fmt.Errorf("failed to write temp build program: %w", err)
	}

	// Change to project directory to run go run from there
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(absDir); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}
	defer os.Chdir(originalDir)

	// Run the temporary build program (use -mod=mod to allow building from temp directory)
	cmd := exec.Command("go", "run", "-mod=mod", tempFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "Alloy_ENV=production")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Now build the final binary
	fmt.Println("📦 Building production binary...")

	outputPath := output
	if outputPath == "" {
		outputPath = filepath.Join(absDir, "dist", "app")
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate temporary app program for production binary
	appProgram := generateAppProgram(moduleName)

	// Create a temporary directory for the app program
	tempAppDir, err := ioutil.TempDir("", "alloy-app-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory for app: %w", err)
	}
	defer os.RemoveAll(tempAppDir)

	// Write the temporary app program
	tempAppFile := filepath.Join(tempAppDir, "main.go")
	if err := os.WriteFile(tempAppFile, []byte(appProgram), 0644); err != nil {
		return fmt.Errorf("failed to write temp app program: %w", err)
	}

	// The build must run from the project directory to properly handle //go:embed directives
	buildCmd := exec.Command("go", "build",
		"-ldflags=-s -w",
		"-o", outputPath,
		"-mod=mod",
		tempAppFile)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Dir = absDir
	buildCmd.Env = append(os.Environ(), "Alloy_ENV=production")

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("binary build failed: %w", err)
	}

	fmt.Printf("✓ Production binary built: %s\n", outputPath)
	return nil
}

func generateBuildProgram(moduleName string) string {
	return fmt.Sprintf(`package main

import (
	"embed"
	"github.com/bertilxi/alloy"
	"github.com/bertilxi/alloy/cli"
	"%s/pages"
)

//go:embed .alloy
var EmbedFS embed.FS

func main() {
	options := alloy.Options{
		EmbedFS:  &EmbedFS,
		Title:    "My Alloy App",
		Loaders:  pages.LoaderRegistry,
		Handlers: pages.HandlerRegistry,
	}
	if err := cli.Build(alloy.New(options)); err != nil {
		panic(err)
	}
}
`, moduleName)
}

func generateAppProgram(moduleName string) string {
	return fmt.Sprintf(`package main

import (
	"embed"
	"github.com/bertilxi/alloy"
	"%s/pages"
)

//go:embed .alloy
var EmbedFS embed.FS

func main() {
	alloy.SetProduction(true)
	options := alloy.Options{
		EmbedFS:  &EmbedFS,
		Title:    "My Alloy App",
		Loaders:  pages.LoaderRegistry,
		Handlers: pages.HandlerRegistry,
	}
	engine := alloy.New(options)
	engine.Start()
}
`, moduleName)
}
