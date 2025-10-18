package cli

import (
	"fmt"
	"os"

	"github.com/bertilxi/htgo"
)

func Build(engine *htgo.Engine) error {
	os.Setenv("HTGO_ENV", string(htgo.HtgoEnvProd))

	PrintBuildStart(engine)

	validationErrors, warnings := ValidatePages(engine)
	if err := PrintValidationResults(validationErrors, warnings); err != nil {
		PrintBuildFailed(len(validationErrors), len(engine.Pages))
		return err
	}

	err := htgo.CleanCache()
	if err != nil {
		return fmt.Errorf("failed to clean cache: %w", err)
	}

	failedCount := 0
	for _, page := range engine.Pages {
		page.AssignOptions(engine.Options)

		bundler := bundler{page: &page}

		PrintPageBuildStart(page.Route, page.File)

		_, backendErr := bundler.buildBackend()
		if backendErr != nil {
			PrintPageBuildError(page.Route, page.File, backendErr)
			failedCount++
			continue
		}

		_, _, clientErr := bundler.buildClient()
		if clientErr != nil {
			PrintPageBuildError(page.Route, page.File, clientErr)
			failedCount++
			continue
		}

		PrintPageBuildComplete(page.Route)
	}

	if failedCount > 0 {
		PrintBuildFailed(failedCount, len(engine.Pages))
		return fmt.Errorf("failed to build %d pages", failedCount)
	}

	PrintBuildComplete(len(engine.Pages), warnings)
	return nil
}
