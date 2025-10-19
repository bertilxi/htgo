package alloy

import (
	"github.com/bertilxi/alloy/core"
)

// DiscoverPages finds all pages in pagesDir and attaches loaders.
func DiscoverPages(pagesDir string, loaders map[string]PageLoader) ([]Page, error) {
	pageFiles, err := core.DiscoverPageFiles(pagesDir)
	if err != nil {
		return nil, err
	}

	pages := make([]Page, len(pageFiles))
	for i, pf := range pageFiles {
		page := Page{
			Route:       pf.Route,
			File:        pf.File,
			Interactive: true,
		}

		if loaders != nil {
			if loader, exists := loaders[pf.Route]; exists {
				page.Loader = loader
			}
		}

		pages[i] = page
	}

	return pages, nil
}
