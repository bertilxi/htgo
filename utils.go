package htgo

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const CacheDir = ".htgo"

type HtgoEnv string

const (
	HtgoEnvProd HtgoEnv = "production"
)

func IsProd() bool {
	return os.Getenv("HTGO_ENV") == string(HtgoEnvProd)
}

func IsDev() bool {
	return os.Getenv("HTGO_ENV") == ""
}

func PageCacheKey(page string, extension string) string {
	pageKey := strings.TrimSuffix(page, filepath.Ext(page))
	cacheKey := fmt.Sprintf("%s.%s", pageKey, extension)
	return path.Join(CacheDir, cacheKey)
}

func CleanCache() error {
	entries, err := os.ReadDir(CacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(CacheDir, 0755)
			if err != nil {
				return err
			}
			return os.WriteFile(path.Join(CacheDir, "keep"), []byte(""), 0644)
		}
		return err
	}

	for _, entry := range entries {
		if entry.Name() != "favicon.svg" && entry.Name() != "keep" {
			err := os.RemoveAll(path.Join(CacheDir, entry.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
