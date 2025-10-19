# Auto-Discovery of Loader Functions in HTGO

## Executive Summary

This report explores feasibility and trade-offs for auto-discovering and loading Go loader functions without manual registration. After analyzing Go's capabilities and the current HTGO architecture, I recommend a **build-time code generation approach** as the most pragmatic solution. This avoids runtime complexity while maintaining HTGO's simplicity-first philosophy.

---

## 1. Go Plugin System Analysis

### What is `go/plugin`?

Go's plugin system (`github.com/go-plugins`) allows runtime loading of compiled `.so` files as plugins. This is **NOT viable for HTGO** for several critical reasons:

#### Critical Constraints

**Platform Support (BLOCKING)**
- Only works on **Linux and macOS** (not Windows, not WebAssembly)
- HTGO targets broad platform support with single-binary deployments
- Plugin system breaks the "single executable" deployment model

**Compilation Requirements**
- Requires building plugins with matching Go toolchain version
- Plugins must be built separately from the main binary
- Can't be baked into production binary via `//go:embed`
- Adds deployment complexity (extra files to ship)

**ABI Instability**
- Go plugin ABI is version-sensitive and platform-specific
- Minor Go version upgrades can break plugins
- No semantic versioning support
- Risk of runtime panics from incompatible plugins

**Runtime Safety**
- Plugin crashes crash the entire application
- No sandboxing or error isolation
- Limited debugging capabilities
- All plugins share same memory space

**Example of the Problem**
```go
// Plugin loading (doesn't work for HTGO's use case)
plugin, err := plugin.Open("loader.so")  // Extra .so file needed
if err != nil {                          // Runtime loading error
    panic(err)
}

sym, err := plugin.Lookup("LoadIndex")
if err != nil {
    panic(err)                           // Panic crashes app
}

loader := sym.(func(*gin.Context) (any, error))
```

### Verdict on go/plugin

**Not Viable**: Plugin system fundamentally conflicts with HTGO's design goals of simplicity and single-binary deployment.

---

## 2. Current HTGO Architecture

### Existing Patterns in Codebase

The codebase already has **partial discovery infrastructure**:

#### Static Code Analysis (in `router.go`)

```go
// ListLoaderFiles finds .go files colocated with pages
func ListLoaderFiles(pagesDir string) ([]string, error)

// Validates loader function signatures via AST parsing
func isValidLoaderSignature(funcDecl *ast.FuncDecl) bool
```

This proves HTGO already uses Go's `go/ast` package for analyzing code structure!

**Key Functions Already Present:**
- `isGinContextType()` - validates parameter types
- `isAnyType()` - validates return types
- `isErrorType()` - validates return types
- Comments note: "Actual loader functions must be registered manually in the Loaders map"

#### Current Manual Registration Pattern

```go
// examples/sink/app.go
Loaders: map[string]func(c *gin.Context) (any, error){
    "/":      pages.LoadIndex,
    "/about": pages.LoadAbout,
}
```

Each page has:
- A `.tsx` component file (e.g., `pages/index.tsx`)
- A `.go` loader file (e.g., `pages/index.go`)
- A manually registered function (e.g., `pages.LoadIndex`)

### Why Manual Registration Was Chosen

1. **Explicitness**: Function names don't have to follow naming conventions
2. **Flexibility**: Can selectively enable/disable loaders
3. **Type Safety**: Caught at compile time, not runtime
4. **Simplicity**: No runtime reflection or code generation needed

---

## 3. Reflection-Based Runtime Discovery

### Approach: Scan Packages at Runtime

Use `reflect` package to discover and register loader functions dynamically.

#### Pseudo-Code Example

```go
// Reflect-based discovery (ANTIPATTERN for HTGO)
func DiscoverLoadersReflection(pkg interface{}) map[string]func(*gin.Context) (any, error) {
    loaders := make(map[string]func(*gin.Context) (any, error))
    
    // Use reflection to scan exported functions
    t := reflect.TypeOf(pkg)
    for i := 0; i < t.NumMethod(); i++ {
        method := t.Method(i)
        
        // Check if signature matches loader pattern
        if isValidSignature(method.Type) {
            loaders[getRoute(method.Name)] = method.Func.Interface().(func(*gin.Context) (any, error))
        }
    }
    
    return loaders
}
```

#### Issues with Reflection

**Cannot Work in Current Architecture**
- Loaders are functions in a `pages` package, not methods on a type
- Reflection works on types/methods, not bare package-level functions
- Would require restructuring pages as a struct with methods

**Runtime Overhead**
- Reflection is slow (1-2 orders of magnitude slower than direct calls)
- Not appropriate for initialization, but startup overhead is minor

**Limited Type Information**
- Reflection provides runtime type info, but loses compile-time safety
- Typos in function names are caught at runtime, not build time
- No IDE autocomplete support

**Example Limitation**
```go
package pages

// Can't reflect on bare functions like this:
func LoadIndex(c *gin.Context) (any, error) { ... }
func LoadAbout(c *gin.Context) (any, error) { ... }

// Would need to refactor to methods on a type:
type Loaders struct{}
func (l Loaders) Index(c *gin.Context) (any, error) { ... }  // Now reflectable

// This adds structure that contradicts HTGO's simplicity
```

### Verdict on Reflection

**Technically Possible But Problematic**: Would require architectural changes and loses HTGO's compile-time safety.

---

## 4. Build-Time Code Generation (RECOMMENDED)

### Approach: Generate loader registration at build time

Use `go:generate` to create a file that registers all discovered loaders automatically.

#### How It Works

```
Step 1: Developer writes loader functions
        pages/index.go:  func LoadIndex(...)  ✓ Named following convention
        pages/about.go:  func LoadAbout(...)  ✓ Named following convention

Step 2: Run `go generate ./...`
        Custom tool scans for loader functions
        Generates: pages/loaders_generated.go

Step 3: Generated file contains:
        var LoaderRegistry = map[string]func(...){
            "/":      pages.LoadIndex,
            "/about": pages.LoadAbout,
        }

Step 4: App uses generated registry:
        Options{
            Loaders: pages.LoaderRegistry,  // Auto-populated!
        }
```

#### Implementation Example

**1. Convention: Loader Naming**

Loaders follow a simple naming pattern based on file path:

```
pages/index.go           → LoadIndex
pages/about.go           → LoadAbout
pages/blog/[slug].go     → LoadBlogSlug (brackets converted to camelCase)
pages/admin/users.go     → LoadAdminUsers
```

**2. Generate Tool** (`cmd/generate-loaders/main.go`)

```go
package main

import (
    "fmt"
    "go/ast"
    "go/parser"
    "os"
    "path/filepath"
    "strings"
)

func main() {
    pagesDir := flag.String("pages", "./pages", "pages directory")
    flag.Parse()

    loaders := discoverLoaders(*pagesDir)
    generated := generateLoaderRegistry(loaders)
    
    outputFile := filepath.Join(*pagesDir, "loaders_generated.go")
    os.WriteFile(outputFile, []byte(generated), 0644)
}

func discoverLoaders(pagesDir string) map[string]string {
    // Same AST parsing already in router.go
    // Returns map[route]functionName
}
```

**3. Generated Output** (`pages/loaders_generated.go`)

```go
// Code generated by htgo generate; DO NOT EDIT

package pages

import "github.com/gin-gonic/gin"

var LoaderRegistry = map[string]func(c *gin.Context) (any, error){
    "/":           LoadIndex,
    "/about":      LoadAbout,
    "/blog/:slug": LoadBlogSlug,
}
```

**4. Application Usage** (`app.go`)

```go
func NewOptions(r *gin.Engine) htgo.Options {
    return htgo.Options{
        Router:   r,
        PagesDir: "./pages",
        Loaders:  pages.LoaderRegistry,  // Auto-populated by code gen
    }
}
```

#### Advantages

1. **No Runtime Overhead**: Generation happens at build time
2. **Type Safe**: Generated code is checked by Go compiler
3. **IDE Friendly**: All imports and references work normally
4. **Debuggable**: Can inspect generated code
5. **Reversible**: Easy to manually edit if needed
6. **Single Binary**: No external files needed
7. **Aligns with HTGO Philosophy**: Explicit, boring, simple

#### Integration with HTGO CLI

```bash
# Option 1: Make `htgo dev/build` run generate automatically
$ htgo dev
→ Runs `go generate ./...` before building
→ Updates loaders_generated.go
→ Starts dev server

# Option 2: Make `htgo install` install the generator
$ htgo install
→ Installs generator tool
→ Shows command to run manually

# Option 3: Standalone generator (simplest)
$ htgo-gen-loaders --pages ./pages
→ Generates pages/loaders_generated.go
```

---

## 5. Comparison Matrix

| Approach | Runtime | Platform | Deployment | Type Safety | Simplicity | Maintenance |
|----------|---------|----------|------------|-------------|-----------|------------|
| **Plugins** | None | ❌ Linux/Mac only | Complex | Partial | Low | Nightmare |
| **Reflection** | Slow | ✓ All | Easy | Low | Medium | Medium |
| **Code Gen** | ✓ None | ✓ All | Easy | High | High | ✓ High |
| **Manual** (current) | ✓ None | ✓ All | Easy | ✓ High | ✓ High | ✓ High |

---

## 6. Trade-offs Analysis

### Build-Time Code Generation vs Current Manual Approach

#### What We Gain
- **Zero Boilerplate**: No need to manually list loaders
- **Auto-Discovery**: New loaders work automatically (after regeneration)
- **Convention Over Config**: Follow naming pattern, everything works
- **Deterministic**: Generated code is reproducible and reviewable

#### What We Lose
- **Explicitness**: Have to understand naming conventions
- **Flexibility**: Can't have loaders with non-conventional names
- **Manual Override**: Harder to selectively disable a loader
- **One Extra Build Step**: Need to run `go generate` before building

#### Middle Ground: Hybrid Approach

**Keep Manual Registration, Add Optional Auto-Gen**

```go
// Developer can write either:

// Option A: Manual (explicit, but tedious)
Loaders: map[string]func(c *gin.Context) (any, error){
    "/": pages.LoadIndex,
}

// Option B: Auto-generated (convenient, follows convention)
Loaders: pages.LoaderRegistry

// Option C: Manual override of auto-generated (best of both)
Loaders: mergeLoaders(pages.LoaderRegistry, map[string]...{
    "/special": pages.CustomSpecialLoader,
})
```

---

## 7. Alternative: Annotation-Based Registration

### Approach: Code Comments as Markers

Use special comments to mark loader functions:

```go
// pages/index.go
package pages

//htgo:loader route="/"
func LoadIndex(c *gin.Context) (any, error) {
    return map[string]any{"data": "..."}, nil
}
```

Generator scans for `//htgo:loader` comments and extracts metadata.

#### Pros
- Self-documenting
- Explicit route mapping in source code
- Flexible (can attach metadata)

#### Cons
- More verbose than naming convention
- Another thing to remember and document
- Duplicate information (file path already implies route)

---

## 8. Recommended Solution

### **Build-Time Code Generation with Convention**

#### Why This is Best

1. **Aligns with HTGO Philosophy**
   - Boring and simple
   - No runtime magic
   - Explicit and debuggable

2. **Minimal Scope Change**
   - Reuses AST parsing already in codebase
   - Leverages existing `go:generate` Go ecosystem
   - Works with current architecture

3. **Zero Runtime Cost**
   - All work happens at build time
   - No reflection or dynamic lookups
   - Single binary deployment

4. **Predictable Behavior**
   - Generated code is reviewable
   - Errors caught by compiler, not runtime
   - Easy to understand what's happening

5. **Gradual Adoption**
   - Can keep manual registration as escape hatch
   - Generate registry alongside manual map
   - Merge or override as needed

#### Implementation Roadmap

**Phase 1: Code Generator Tool**
- Extract AST parsing from `router.go` into standalone tool
- Create generator: `cmd/htgo-gen-loaders`
- Can be run independently: `go run cmd/htgo-gen-loaders --pages ./pages`

**Phase 2: Integration with CLI**
- Add `htgo generate` command
- Make `htgo dev` and `htgo build` run generate automatically
- Update docs with naming convention

**Phase 3: Optional Enhancement**
- Add `--manual-only` flag for explicit projects
- Support manual overrides in generated file
- Add IDE plugin/tooling to help with naming

---

## 9. Naming Convention Details

### File Path to Function Name Mapping

```
File Path                  →  Function Name    →  Route
pages/index.go            →  LoadIndex        →  /
pages/about.go            →  LoadAbout        →  /about
pages/blog.go             →  LoadBlog         →  /blog
pages/blog/[slug].go      →  LoadBlogSlug     →  /blog/:slug
pages/api/v1/users.go     →  LoadApiV1Users   →  /api/v1/users
pages/user_profile.go     →  LoadUserProfile  →  /user-profile
pages/admin/[...id].go    →  LoadAdminId      →  /admin/:id*
```

### Convention Rules

1. Start with `Load` prefix
2. Remove `.go` extension
3. Replace directory separators with camelCase
4. Replace brackets with camelCase (e.g., `[slug]` → `Slug`)
5. Replace underscores with camelCase (e.g., `user_profile` → `UserProfile`)
6. Replace dots with underscores (e.g., `v1.0` → `v1_0`)

---

## 10. Implementation Checklist

### For HTGO Core

- [ ] Extract AST parsing utilities to reusable package
- [ ] Create standalone `htgo-gen-loaders` tool
- [ ] Add to CLAUDE.md: Loader naming convention
- [ ] Update router.go to use generated registry (when available)
- [ ] Add `--skip-generate` flag for manual override

### For Examples

- [ ] Update `examples/sink/pages/loaders_generated.go`
- [ ] Update `examples/sink/app.go` to use `LoaderRegistry`
- [ ] Remove manual registration from examples
- [ ] Add `Makefile` or shell script to show `go generate` usage

### For Documentation

- [ ] Document naming convention
- [ ] Show generated output example
- [ ] Explain how to run generator
- [ ] Show how to override generated registry

---

## 11. Risks and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Developer forgets to run generate | High | New loaders silently ignored | Integrate into `htgo dev/build` commands |
| Generated file conflicts with manual edits | Low | Confusion/merge issues | Use clear `DO NOT EDIT` comments; support .gitignore |
| Naming convention too complex | Medium | Developer errors | Simple, clear rules; good error messages |
| Performance impact | None | N/A | No runtime overhead (build time) |

---

## 12. Conclusion

**Recommendation: Implement build-time code generation**

This approach:
- Maintains HTGO's simplicity-first philosophy
- Requires no architectural changes
- Has zero runtime overhead
- Provides better developer experience than manual registration
- Is reversible if needed

The naming convention is straightforward, and the generated code is transparent and reviewable. This is a boring, pragmatic solution that fits HTGO's design principles perfectly.

### Why Not The Others?

- **Plugins**: Breaks single-binary deployment, platform limitations
- **Reflection**: Requires architectural changes, loses compile-time safety
- **Manual**: Current approach is fine but generates boilerplate

---

## Appendix: Code Examples

### Example Generated Registry

```go
// Code generated by htgo generate; DO NOT EDIT

package pages

import "github.com/gin-gonic/gin"

// LoaderRegistry maps routes to their handler functions.
// This file is auto-generated from loader functions in this package.
// To regenerate: go generate ./...
var LoaderRegistry = map[string]func(c *gin.Context) (any, error){
	"/":           LoadIndex,
	"/about":      LoadAbout,
	"/blog":       LoadBlog,
	"/blog/:slug": LoadBlogSlug,
}
```

### Example Generator Usage

```bash
# Manual usage
go run ./cmd/htgo-gen-loaders -pages ./pages -output ./pages/loaders_generated.go

# Via go:generate (in package.go)
//go:generate go run ../../cmd/htgo-gen-loaders -pages . -output ./loaders_generated.go

# Via htgo CLI
htgo generate
htgo dev  # Automatically runs generate
htgo build # Automatically runs generate
```

### Example Project Structure After Implementation

```
sink/
├── app.go                    # Uses pages.LoaderRegistry
├── app/
│   └── pages/
│       ├── index.go          # func LoadIndex
│       ├── index.tsx         # Component
│       ├── about.go          # func LoadAbout
│       ├── about.tsx         # Component
│       ├── loaders_generated.go  # GENERATED by htgo
│       └── package.go        # Contains //go:generate directive
└── cmd/
    └── app/main.go
```

