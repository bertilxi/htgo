package cli

import (
	"os"

	"github.com/bertilxi/htgo"
)

func Build(engine *htgo.Engine) error {
	os.Setenv("HTGO_ENV", string(htgo.HtgoEnvProd))

	err := htgo.CleanCache()
	if err != nil {
		return err
	}

	for _, page := range engine.Pages {
		page.AssignOptions(engine.Options)

		bundler := bundler{page: &page}
		bundler.buildBackend()
		bundler.buildClient()
	}

	return nil
}
