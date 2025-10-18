package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bertilxi/htgo"
	"github.com/bertilxi/htgo/cli"
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

	appFilePath := filepath.Join(absDir, "app.go")
	if _, err := os.Stat(appFilePath); err != nil {
		return fmt.Errorf("app.go not found in %s - are you in an HTGO project?", dir)
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

func loadEngine(dir string) (*htgo.Engine, error) {
	fmt.Println("⚠️  Note: Dynamic project loading not yet fully implemented.")
	fmt.Println("Please run: cd <project> && go run cmd/dev/main.go")
	fmt.Println("")
	fmt.Println("Full CLI integration coming soon!")
	os.Exit(1)
	return nil, nil
}
