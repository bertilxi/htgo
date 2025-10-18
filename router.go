package htgo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func DiscoverPages(pagesDir string, loaders map[string]func(c *gin.Context) (any, error)) ([]Page, error) {
	if pagesDir == "" {
		return nil, fmt.Errorf("pagesDir is required")
	}

	absPageDir, err := filepath.Abs(pagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for pagesDir: %w", err)
	}

	var pages []Page

	err = filepath.Walk(absPageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".tsx") {
			return nil
		}

		route := filePathToRoute(path, absPageDir)

		relPath, err := filepath.Rel(absPageDir, path)
		if err != nil {
			relPath = path
		}
		relPath = filepath.Join(pagesDir, relPath)

		page := Page{
			Route:       route,
			File:        relPath,
			Interactive: true,
		}

		if loaders != nil {
			if handler, exists := loaders[route]; exists {
				page.Handler = handler
			}
		}

		pages = append(pages, page)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover pages: %w", err)
	}

	return pages, nil
}

func filePathToRoute(filePath string, pagesDir string) string {
	relativePath := strings.TrimPrefix(filePath, pagesDir)
	relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))
	relativePath = strings.TrimSuffix(relativePath, ".tsx")

	if relativePath == "index" {
		return "/"
	}

	fileParts := strings.Split(relativePath, string(filepath.Separator))
	routeParts := make([]string, len(fileParts))

	for i, part := range fileParts {
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			paramName := strings.TrimPrefix(part, "[")
			paramName = strings.TrimSuffix(paramName, "]")
			routeParts[i] = ":" + paramName
		} else {
			routeParts[i] = part
		}
	}

	return "/" + strings.Join(routeParts, "/")
}

// ListLoaderFiles finds .go files colocated with pages for documentation purposes
// This helps identify which loader files exist, but actual loader functions
// must be registered manually in the Loaders map
func ListLoaderFiles(pagesDir string) ([]string, error) {
	if pagesDir == "" {
		return nil, fmt.Errorf("pagesDir is required")
	}

	absPageDir, err := filepath.Abs(pagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for pagesDir: %w", err)
	}

	var loaderFiles []string

	err = filepath.Walk(absPageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Check if there's a corresponding .tsx file
		tsxPath := strings.TrimSuffix(path, ".go") + ".tsx"
		if _, err := os.Stat(tsxPath); err != nil {
			// No corresponding .tsx file, skip
			return nil
		}

		// Verify it has a valid loader function signature
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		for _, decl := range node.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			// Must be exported (starts with uppercase)
			if !funcDecl.Name.IsExported() {
				continue
			}

			// Check if it matches the loader signature
			if isValidLoaderSignature(funcDecl) {
				relPath, _ := filepath.Rel(absPageDir, path)
				loaderFiles = append(loaderFiles, relPath)
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list loader files: %w", err)
	}

	return loaderFiles, nil
}

// isValidLoaderSignature checks if a function has the loader signature
// func(c *gin.Context) (any, error)
func isValidLoaderSignature(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil || funcDecl.Type.Results == nil {
		return false
	}

	// Check parameters: should have exactly 1 param of type *gin.Context
	if len(funcDecl.Type.Params.List) != 1 {
		return false
	}

	param := funcDecl.Type.Params.List[0]
	if !isGinContextType(param.Type) {
		return false
	}

	// Check return types: should be (any, error)
	if len(funcDecl.Type.Results.List) != 2 {
		return false
	}

	// First return should be 'any'
	if !isAnyType(funcDecl.Type.Results.List[0].Type) {
		return false
	}

	// Second return should be 'error'
	if !isErrorType(funcDecl.Type.Results.List[1].Type) {
		return false
	}

	return true
}

func isGinContextType(expr ast.Expr) bool {
	starExpr, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}

	selExpr, ok := starExpr.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "gin" && selExpr.Sel.Name == "Context"
}

func isAnyType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "any"
}

func isErrorType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "error"
}
