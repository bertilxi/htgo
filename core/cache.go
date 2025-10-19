package core

import (
	"embed"
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

type BundleReader interface {
	ReadBundle(name string) ([]byte, error)
}

type FileSystemBundleReader struct {
	Dev     bool
	EmbedFS *embed.FS
}

func (r *FileSystemBundleReader) ReadBundle(name string) ([]byte, error) {
	if r.Dev || r.EmbedFS == nil {
		return os.ReadFile(name)
	}
	return r.EmbedFS.ReadFile(name)
}

func GetServerBundle(reader BundleReader, cacheKey string) (string, error) {
	if val, ok := bundleCache.Load(cacheKey); ok {
		return val.(string), nil
	}

	cached, err := reader.ReadBundle(cacheKey)
	if err != nil {
		return "", err
	}

	bundle := string(cached)
	bundleCache.Store(cacheKey, bundle)
	return bundle, nil
}

func GetClientBundles(reader BundleReader, jsKey, cssKey string) (string, string, error) {
	_, jsErr := reader.ReadBundle(jsKey)
	_, cssErr := reader.ReadBundle(cssKey)
	if jsErr != nil || cssErr != nil {
		return "", "", jsErr
	}

	return jsKey, cssKey, nil
}
