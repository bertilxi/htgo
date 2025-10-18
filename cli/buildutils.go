package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bertilxi/htgo"
)

type BuildStats struct {
	TotalPages      int
	SuccessCount    int
	FailureCount    int
	Warnings        []string
	TotalBundleSize int64
}

type PageError struct {
	Page  string
	Error string
	File  string
}

func ValidatePages(engine *htgo.Engine) ([]PageError, []string) {
	var errors []PageError
	var warnings []string

	for _, page := range engine.Pages {
		if page.File == "" {
			errors = append(errors, PageError{
				Page:  page.Route,
				Error: "Page.File is empty",
				File:  "",
			})
			continue
		}

		fileInfo, err := os.Stat(page.File)
		if err != nil {
			errors = append(errors, PageError{
				Page:  page.Route,
				Error: fmt.Sprintf("File not found: %s", err),
				File:  page.File,
			})
			continue
		}

		ext := filepath.Ext(page.File)
		if ext != ".tsx" && ext != ".jsx" && ext != ".ts" && ext != ".js" {
			errors = append(errors, PageError{
				Page:  page.Route,
				Error: fmt.Sprintf("Invalid file extension '%s'. Expected .tsx, .jsx, .ts, or .js", ext),
				File:  page.File,
			})
			continue
		}

		if fileInfo.IsDir() {
			errors = append(errors, PageError{
				Page:  page.Route,
				Error: "Path is a directory, not a file",
				File:  page.File,
			})
			continue
		}

		if fileInfo.Size() == 0 {
			warnings = append(warnings, fmt.Sprintf("⚠️  Page %s (%s) is empty", page.Route, page.File))
		}

		if !strings.Contains(page.File, "export") {
			content, _ := os.ReadFile(page.File)
			if !strings.Contains(string(content), "export") {
				warnings = append(warnings, fmt.Sprintf("⚠️  Page %s might not export a component", page.Route))
			}
		}
	}

	return errors, warnings
}

func PrintValidationResults(errors []PageError, warnings []string) error {
	if len(errors) > 0 {
		fmt.Println()
		fmt.Println("❌ Build validation failed:")
		fmt.Println()
		for i, err := range errors {
			fmt.Printf("  %d. Route '%s':\n", i+1, err.Page)
			fmt.Printf("     Error: %s\n", err.Error)
			if err.File != "" {
				fmt.Printf("     File: %s\n", err.File)
			}
			fmt.Println()
		}
		return fmt.Errorf("%d validation errors found", len(errors))
	}

	if len(warnings) > 0 {
		fmt.Println()
		for _, warning := range warnings {
			fmt.Println(warning)
		}
		fmt.Println()
	}

	return nil
}

func PrintBuildStart(engine *htgo.Engine) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📦 Starting Production Build")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📄 Pages to bundle: %d\n", len(engine.Pages))
	for _, page := range engine.Pages {
		fmt.Printf("   • %s → %s\n", page.Route, page.File)
	}
	fmt.Println()
}

func PrintPageBuildStart(route, file string) {
	fmt.Printf("📌 Bundling %s (%s)...\n", route, file)
}

func PrintPageBuildComplete(route string) {
	fmt.Printf("✓ %s bundled\n", route)
}

func PrintPageBuildError(route, file string, err error) {
	fmt.Printf("❌ %s failed: %v\n", route, err)
}

func PrintBuildComplete(totalPages int, warnings []string) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✓ Production Build Complete")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📦 Successfully bundled %d pages\n", totalPages)
	if len(warnings) > 0 {
		fmt.Printf("⚠️  %d warnings\n", len(warnings))
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Run production server: htgo start")
	fmt.Println("  • Or run in development: htgo dev")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

func PrintBuildFailed(failedCount int, totalCount int) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("❌ Build failed: %d of %d pages could not be bundled\n", failedCount, totalCount)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

func ExtractBuildErrorContext(errMsg string) string {
	errMsg = strings.TrimSpace(errMsg)

	if strings.Contains(errMsg, "Cannot find module") {
		return "Import error: Check that imported modules exist and are installed"
	}
	if strings.Contains(errMsg, "Module not found") {
		return "Module import error: Check npm dependencies"
	}
	if strings.Contains(errMsg, "SyntaxError") {
		return "TypeScript/JSX syntax error: Check component syntax"
	}
	if strings.Contains(errMsg, "Unexpected token") {
		return "Parsing error: Invalid syntax in component"
	}
	if strings.Contains(errMsg, "Invalid JSX") {
		return "Invalid JSX: Check component JSX syntax"
	}

	if len(errMsg) > 150 {
		return errMsg[:150] + "..."
	}

	return errMsg
}
