package cli

import (
	"fmt"
	"os"
	"sync"

	"github.com/bertilxi/htgo"
)

func Build(engine *htgo.Engine) error {
	os.Setenv("HTGO_ENV", string(htgo.HtgoEnvProd))

	PrintBuildStart(engine)

	// Generate loader registry from .go files
	err := ensureGeneratedLoaders(engine.Options.PagesDir)
	if err != nil {
		return err
	}

	// Discover pages
	pages, err := htgo.DiscoverPages(engine.Options.PagesDir, engine.Options.Handlers)
	if err != nil {
		return err
	}
	engine.Pages = pages

	// Ensure Tailwind is available before building pages
	err = EnsureTailwind(engine.Pages)
	if err != nil {
		return err
	}

	validationErrors, warnings := ValidatePages(engine)
	if err := PrintValidationResults(validationErrors, warnings); err != nil {
		PrintBuildFailed(len(validationErrors), len(engine.Pages))
		return err
	}

	err = htgo.CleanCache()
	if err != nil {
		return fmt.Errorf("failed to clean cache: %w", err)
	}

	type buildResult struct {
		page  htgo.Page
		err   error
	}

	resultsCh := make(chan buildResult, len(engine.Pages))
	var wg sync.WaitGroup

	for _, page := range engine.Pages {
		wg.Add(1)
		go func(p htgo.Page) {
			defer wg.Done()
			p.AssignOptions(engine.Options)
			bundler := bundler{page: &p}

			PrintPageBuildStart(p.Route, p.File)

			_, backendErr := bundler.buildBackend()
			if backendErr != nil {
				PrintPageBuildError(p.Route, p.File, backendErr)
				resultsCh <- buildResult{page: p, err: backendErr}
				return
			}

			_, _, clientErr := bundler.buildClient()
			if clientErr != nil {
				PrintPageBuildError(p.Route, p.File, clientErr)
				resultsCh <- buildResult{page: p, err: clientErr}
				return
			}

			PrintPageBuildComplete(p.Route)
			resultsCh <- buildResult{page: p, err: nil}
		}(page)
	}

	wg.Wait()
	close(resultsCh)

	failedCount := 0
	for result := range resultsCh {
		if result.err != nil {
			failedCount++
		}
	}

	if failedCount > 0 {
		PrintBuildFailed(failedCount, len(engine.Pages))
		return fmt.Errorf("failed to build %d pages", failedCount);
	}

	PrintBuildComplete(len(engine.Pages), warnings)
	return nil
}
