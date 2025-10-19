// Package loaderutil provides utilities for discovering and validating loader functions.
package loaderutil

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// LoaderInfo represents a discovered loader function
type LoaderInfo struct {
	Route        string // e.g., "/", "/about", "/blog/:slug"
	FunctionName string // e.g., "LoadIndex", "LoadAbout"
	FilePath     string // relative path to .go file, e.g., "pages/index.go"
	IsAPI        bool   // true if this is an API handler (in pages/api/), false if page loader
}

// DiscoverLoaders finds all .go files with valid loader and API handler functions in pagesDir
func DiscoverLoaders(pagesDir string) ([]LoaderInfo, error) {
	if pagesDir == "" {
		return nil, fmt.Errorf("pagesDir is required")
	}

	absPageDir, err := filepath.Abs(pagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for pagesDir: %w", err)
	}

	var loaders []LoaderInfo

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

		// Check if this is in the api subdirectory
		isAPIFile := strings.Contains(path, filepath.Join(absPageDir, "api"))

		if !isAPIFile {
			// For page loaders: check if there's a corresponding .tsx file
			tsxPath := strings.TrimSuffix(path, ".go") + ".tsx"
			if _, err := os.Stat(tsxPath); err != nil {
				// No corresponding .tsx file, skip
				return nil
			}
		}

		// Parse the Go file to find exported functions
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		// Look for the first exported function matching loader or API handler signature
		for _, decl := range node.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			// Must be exported (starts with uppercase)
			if !funcDecl.Name.IsExported() {
				continue
			}

			if isAPIFile {
				// Check if it matches the API handler signature: func(c *gin.Context)
				if IsValidAPIHandlerSignature(funcDecl) {
					route := FilePathToRoute(path, absPageDir, true)
					relPath, _ := filepath.Rel(absPageDir, path)

					loaders = append(loaders, LoaderInfo{
						Route:        route,
						FunctionName: funcDecl.Name.Name,
						FilePath:     relPath,
						IsAPI:        true,
					})
					break
				}
			} else {
				// Check if it matches the loader signature: func(c *gin.Context) (any, error)
				if IsValidLoaderSignature(funcDecl) {
					route := FilePathToRoute(path, absPageDir, false)
					relPath, _ := filepath.Rel(absPageDir, path)

					loaders = append(loaders, LoaderInfo{
						Route:        route,
						FunctionName: funcDecl.Name.Name,
						FilePath:     relPath,
						IsAPI:        false,
					})
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover loaders: %w", err)
	}

	return loaders, nil
}

// IsValidLoaderSignature checks if a function has the loader signature:
// func(c *gin.Context) (any, error)
func IsValidLoaderSignature(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil || funcDecl.Type.Results == nil {
		return false
	}

	// Check parameters: should have exactly 1 param of type *gin.Context
	if len(funcDecl.Type.Params.List) != 1 {
		return false
	}

	param := funcDecl.Type.Params.List[0]
	if !IsGinContextType(param.Type) {
		return false
	}

	// Check return types: should be (any, error)
	if len(funcDecl.Type.Results.List) != 2 {
		return false
	}

	// First return should be 'any'
	if !IsAnyType(funcDecl.Type.Results.List[0].Type) {
		return false
	}

	// Second return should be 'error'
	if !IsErrorType(funcDecl.Type.Results.List[1].Type) {
		return false
	}

	return true
}

// IsGinContextType checks if expr is *gin.Context
func IsGinContextType(expr ast.Expr) bool {
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

// IsAnyType checks if expr is 'any'
func IsAnyType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "any"
}

// IsErrorType checks if expr is 'error'
func IsErrorType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "error"
}

// IsValidAPIHandlerSignature checks if a function has the unified handler signature:
// func(c *gin.Context) (any, error)
// This is the same signature as page loaders, allowing flexibility in return types.
func IsValidAPIHandlerSignature(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil || funcDecl.Type.Results == nil {
		return false
	}

	// Check parameters: should have exactly 1 param of type *gin.Context
	if len(funcDecl.Type.Params.List) != 1 {
		return false
	}

	param := funcDecl.Type.Params.List[0]
	if !IsGinContextType(param.Type) {
		return false
	}

	// Check return types: should be (any, error)
	if len(funcDecl.Type.Results.List) != 2 {
		return false
	}

	// First return should be 'any'
	if !IsAnyType(funcDecl.Type.Results.List[0].Type) {
		return false
	}

	// Second return should be 'error'
	if !IsErrorType(funcDecl.Type.Results.List[1].Type) {
		return false
	}

	return true
}

// FilePathToRoute converts a .go file path to its route
// For API handlers, adds /api prefix. For page loaders, no prefix.
// e.g., "pages/index.go" -> "/", "pages/about.go" -> "/about", "pages/api/hello.go" -> "/api/hello"
func FilePathToRoute(filePath string, pagesDir string, isAPI bool) string {
	relativePath := strings.TrimPrefix(filePath, pagesDir)
	relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))
	relativePath = strings.TrimSuffix(relativePath, ".go")

	// For API handlers in api/ subdirectory, remove "api/" prefix since we'll add it back
	if isAPI && strings.HasPrefix(relativePath, "api"+string(filepath.Separator)) {
		relativePath = strings.TrimPrefix(relativePath, "api"+string(filepath.Separator))
	}

	// For page loaders, index -> /
	if !isAPI && relativePath == "index" {
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

	route := "/" + strings.Join(routeParts, "/")

	// For API handlers, add /api prefix
	if isAPI {
		route = "/api" + route
	}

	return route
}

// FilePathToFunctionName converts a file path to a function name
// e.g., "index" -> "LoadIndex", "blog/[slug]" -> "LoadBlogSlug"
func FilePathToFunctionName(filePath string) string {
	// Remove directory paths - keep just the filename
	base := filepath.Base(filePath)
	// Remove .go extension
	base = strings.TrimSuffix(base, ".go")
	// Remove .tsx extension if present
	base = strings.TrimSuffix(base, ".tsx")

	// Split by _ and capitalize each part
	parts := strings.Split(base, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		// Capitalize first letter
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}

	// Get the full path and convert each segment
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		// Just the filename
		return "Load" + strings.Join(parts, "")
	}

	// Build function name from directory and file
	segments := strings.Split(filepath.ToSlash(filePath), "/")
	var funcName strings.Builder
	funcName.WriteString("Load")

	for _, segment := range segments[:len(segments)-1] { // all but filename
		if segment == "." || segment == "" {
			continue
		}
		// Handle [param] style directories
		if strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]") {
			param := strings.TrimPrefix(segment, "[")
			param = strings.TrimSuffix(param, "]")
			funcName.WriteString(strings.ToUpper(param[:1]) + param[1:])
		} else {
			funcName.WriteString(strings.ToUpper(segment[:1]) + segment[1:])
		}
	}

	// Add the filename part (without extension)
	filename := strings.TrimSuffix(filepath.Base(filePath), ".go")
	filename = strings.TrimSuffix(filename, ".tsx")
	filenameParts := strings.Split(filename, "_")
	for _, part := range filenameParts {
		if part != "" {
			funcName.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}

	return funcName.String()
}
