package alloy

import (
	"github.com/bertilxi/alloy/core"
)

// ClearBundleCache clears the in-memory bundle cache.
func ClearBundleCache() {
	core.ClearBundleCache()
}

func (page *Page) getBundleReader() *core.FileSystemBundleReader {
	return &core.FileSystemBundleReader{
		Dev:     core.IsDev(),
		EmbedFS: page.embedFS,
	}
}

func (page *Page) getServerJsFromFs() (string, error) {
	cacheKey := core.PageCacheKey(page.File, "ssr.js")
	reader := page.getBundleReader()
	return core.GetServerBundle(reader, cacheKey)
}

func (page *Page) getClientJsFromFs() (string, string, error) {
	jsCacheKey := core.PageCacheKey(page.File, "js")
	cssCacheKey := core.PageCacheKey(page.File, "css")
	reader := page.getBundleReader()
	return core.GetClientBundles(reader, jsCacheKey, cssCacheKey)
}
