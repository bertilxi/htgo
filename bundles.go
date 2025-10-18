package htgo

import (
	"os"
	"sync"
)

var bundleCache sync.Map

func ClearBundleCache() {
	bundleCache.Range(func(key, value interface{}) bool {
		bundleCache.Delete(key)
		return true
	})
}

func (page *Page) readFile(name string) ([]byte, error) {
	if IsDev() || page.embedFS == nil {
		return os.ReadFile(name)
	}

	return page.embedFS.ReadFile(name)
}

func (page *Page) getServerJsFromFs() (string, error) {
	cacheKey := PageCacheKey(page.File, "ssr.js")

	if val, ok := bundleCache.Load(cacheKey); ok {
		return val.(string), nil
	}

	cached, err := page.readFile(cacheKey)
	if err != nil {
		return "", err
	}

	bundle := string(cached)
	bundleCache.Store(cacheKey, bundle)
	return bundle, nil
}

func (page *Page) getClientJsFromFs() (string, string, error) {
	jsCacheKey := PageCacheKey(page.File, "js")
	cssCacheKey := PageCacheKey(page.File, "css")

	_, jsErr := page.readFile(jsCacheKey)
	_, cssErr := page.readFile(cssCacheKey)
	if jsErr != nil || cssErr != nil {
		return "", "", jsErr
	}

	return jsCacheKey, cssCacheKey, nil
}
