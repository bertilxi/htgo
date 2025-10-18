package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bertilxi/htgo/cli"
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

	appFilePath := filepath.Join(absDir, "app.go")
	if _, err := os.Stat(appFilePath); err != nil {
		return fmt.Errorf("app.go not found in %s - are you in an HTGO project?", dir)
	}

	fmt.Printf("📁 Building project from: %s\n", absDir)

	engine, err := loadEngine(absDir)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	fmt.Println("📦 Building bundles...")
	if err := cli.Build(engine); err != nil {
		return fmt.Errorf("bundle build failed: %w", err)
	}

	fmt.Println("✓ Bundles built successfully")

	if output != "" {
		fmt.Printf("📝 Build output would be placed at: %s\n", output)
	}

	fmt.Println("✓ Production build complete")
	return nil
}
