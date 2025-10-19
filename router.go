package alloy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DiscoverPages(pagesDir string, loaders map[string]PageLoader) ([]Page, error) {
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
			if loader, exists := loaders[route]; exists {
				page.Loader = loader
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
