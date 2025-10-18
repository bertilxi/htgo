package htgo

import (
	"encoding/json"
	"testing"
)

func BenchmarkBundleCache(b *testing.B) {
	testBundle := `var React = {}; function renderPage() { return "<div>Test</div>"; };`
	cacheKey := "test.ssr.js"

	bundleCache.Store(cacheKey, testBundle)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bundleCache.Load(cacheKey)
	}
}

func BenchmarkPropsMarshaling(b *testing.B) {
	props := map[string]any{
		"title":    "Test Page",
		"route":    "/test",
		"time":     "2024-10-18T10:00:00Z",
		"items":    []string{"a", "b", "c"},
		"metadata": map[string]string{"key": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(props)
	}
}

func BenchmarkErrorExtraction(b *testing.B) {
	errors := []string{
		"ReferenceError: undefined variable",
		"TypeError: cannot read property",
		"SyntaxError: unexpected token",
		"Cannot read property 'foo' of undefined",
		"Long error message that exceeds two hundred characters and should be truncated to prevent excessive output in error messages displayed to users during development or production debugging sessions",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, err := range errors {
			extractJSErrorContext(err)
		}
	}
}

func BenchmarkAssetURL(b *testing.B) {
	pageContexts["test.tsx"] = pageContext{
		assetURLPrefix:   "/cdn",
		cacheBustVersion: "v1.0.0",
	}
	page := &Page{File: "test.tsx"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page.assetURL("bundle.js")
	}
}
