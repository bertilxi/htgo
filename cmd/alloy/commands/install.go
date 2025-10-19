package commands

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bertilxi/alloy"
	"github.com/bertilxi/alloy/cli"
)

func InstallCmd(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	dir := fs.String("dir", ".", "Project directory")

	fs.Parse(args)

	if err := runInstall(*dir); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Install error: %v\n", err)
		os.Exit(1)
	}
}

func runInstall(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	appFilePath := filepath.Join(absDir, "app.go")
	if _, err := os.Stat(appFilePath); err != nil {
		return fmt.Errorf("app.go not found in %s - are you in an Alloy project?", dir)
	}

	fmt.Printf("📦 Installing dependencies...\n\n")

	// Change to project directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(absDir); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}
	defer os.Chdir(originalDir)

	// 1. Run go mod tidy
	fmt.Println("📦 Running 'go mod tidy'...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}
	fmt.Println("✓ Go dependencies tidied")

	// 2. Run npm install
	fmt.Println("\n📦 Running 'npm install'...")
	cmd = exec.Command("npm", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}
	fmt.Println("✓ NPM packages installed")

	// 3. Create .alloy directory
	fmt.Println("\n📦 Creating .alloy directory...")
	alloyDir := filepath.Join(absDir, ".alloy")
	if err := os.MkdirAll(alloyDir, 0755); err != nil {
		return fmt.Errorf("failed to create .alloy directory: %w", err)
	}
	fmt.Println("✓ .alloy directory created")

	// 4. Create .alloy/keep file to ensure .alloy is tracked by git
	fmt.Println("\n📦 Creating .alloy/keep file...")
	keepFile := filepath.Join(alloyDir, "keep")
	if err := os.WriteFile(keepFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create .alloy/keep file: %w", err)
	}
	fmt.Println("✓ .alloy/keep file created")

	// 5. Check if Tailwind is used and download it if needed
	fmt.Println("\n📦 Checking for Tailwind CSS...")
	if hasTailwindInProject(absDir) {
		// Change to project dir for EnsureTailwind
		origDir, _ := os.Getwd()
		os.Chdir(absDir)
		pagesDir := filepath.Join(absDir, "pages")
		pages, err := alloy.DiscoverPages(pagesDir, nil)
		if err == nil && len(pages) > 0 {
			if err := cli.EnsureTailwind(pages); err != nil {
				os.Chdir(origDir)
				fmt.Printf("⚠️  Warning: Failed to download Tailwind: %v\n", err)
			} else {
				os.Chdir(origDir)
				fmt.Println("✓ Tailwind CSS downloaded and ready")
			}
		} else {
			os.Chdir(origDir)
			fmt.Println("✓ Tailwind CSS (will download on first use if needed)")
		}
	} else {
		fmt.Println("✓ Tailwind CSS (will download on first use if needed)")
	}

	fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✓ Installation complete!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Start development:  alloy dev")
	fmt.Println("  2. Build for production: alloy build")
	fmt.Println("  3. Run production: ./dist/app")

	return nil
}

// hasTailwindInProject checks if any CSS file in the project uses Tailwind
func hasTailwindInProject(dir string) bool {
	// Check root CSS files
	if content, err := os.ReadFile(filepath.Join(dir, "styles.css")); err == nil {
		if strings.Contains(string(content), `@import "tailwindcss"`) {
			return true
		}
	}

	// Check CSS files in pages directory
	pagesDir := filepath.Join(dir, "pages")
	if entries, err := os.ReadDir(pagesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".css") {
				if content, err := os.ReadFile(filepath.Join(pagesDir, entry.Name())); err == nil {
					if strings.Contains(string(content), `@import "tailwindcss"`) {
						return true
					}
				}
			}
		}
	}

	return false
}
