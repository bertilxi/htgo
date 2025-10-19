// Package apiutil provides utilities for discovering and validating API handler functions.
package apiutil

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// APIInfo represents a discovered API handler function
type APIInfo struct {
	Route        string // e.g., "/api/posts", "/api/users/:id"
	FunctionName string // e.g., "PostsHandler", "UsersHandler"
	FilePath     string // relative path to .go file, e.g., "api/posts.go"
}

// DiscoverAPIHandlers finds all .go files with valid API handler functions in apiDir
func DiscoverAPIHandlers(apiDir string) ([]APIInfo, error) {
	if apiDir == "" {
		return nil, fmt.Errorf("apiDir is required")
	}

	absAPIDir, err := filepath.Abs(apiDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for apiDir: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absAPIDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Directory doesn't exist, return empty list
		}
		return nil, fmt.Errorf("failed to access apiDir: %w", err)
	}

	var handlers []APIInfo

	err = filepath.Walk(absAPIDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse the Go file to find exported functions
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		// Look for all exported functions matching API handler signature
		for _, decl := range node.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			// Must be exported (starts with uppercase)
			if !funcDecl.Name.IsExported() {
				continue
			}

			// Check if it matches the API handler signature
			if IsValidAPIHandlerSignature(funcDecl) {
				route := FilePathToRoute(path, absAPIDir)
				relPath, _ := filepath.Rel(absAPIDir, path)

				handlers = append(handlers, APIInfo{
					Route:        route,
					FunctionName: funcDecl.Name.Name,
					FilePath:     relPath,
				})
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover API handlers: %w", err)
	}

	return handlers, nil
}

// IsValidAPIHandlerSignature checks if a function has the API handler signature:
// func(c *gin.Context)
func IsValidAPIHandlerSignature(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil {
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

	// Check return types: should be void (no return values)
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
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

// FilePathToRoute converts a .go file path to its route with /api prefix
// e.g., "api/posts.go" -> "/api/posts", "api/users/[id].go" -> "/api/users/:id"
func FilePathToRoute(filePath string, apiDir string) string {
	relativePath := strings.TrimPrefix(filePath, apiDir)
	relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))
	relativePath = strings.TrimSuffix(relativePath, ".go")

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

	return "/api/" + strings.Join(routeParts, "/")
}
