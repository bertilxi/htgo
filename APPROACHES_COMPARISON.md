# Comprehensive Comparison: Loader Auto-Discovery Approaches

## Executive Decision Matrix

| Criterion | Go Plugins | Reflection | Code Generation | Manual (Current) |
|-----------|-----------|-----------|-----------------|-----------------|
| **Platform Support** | ❌ Linux/macOS only | ✓ All platforms | ✓ All platforms | ✓ All platforms |
| **Single Binary** | ❌ Extra .so files | ✓ Yes | ✓ Yes | ✓ Yes |
| **Type Safety** | Partial | ❌ Low | ✓ High | ✓ High |
| **Compile-time Errors** | ❌ No | ❌ No | ✓ Yes | ✓ Yes |
| **IDE Support** | ✓ Good | ❌ Poor | ✓ Good | ✓ Good |
| **Runtime Overhead** | High | High | None | None |
| **Debuggability** | Hard | Medium | ✓ Easy | ✓ Easy |
| **Learning Curve** | Medium | Medium | ✓ Low | ✓ Low |
| **Maintenance** | Hard | Medium | ✓ Easy | ✓ Easy |
| **Reversibility** | Hard | Easy | ✓ Easy | N/A |
| **Flexibility** | High | High | Medium | ✓ High |
| **HTGO Philosophy** | ❌ No | No | ✓ Yes | ✓ Yes |

---

## Detailed Analysis

### Approach 1: Go Plugins (go/plugin)

#### What It Is
Dynamic loading of compiled `.so` files as plugins at runtime.

#### How It Would Work (Hypothetically)

```
1. Write loader in separate module
   loaders/index.so
   
2. Build plugin separately
   $ go build -buildmode=plugin -o loaders/index.so loaders/index.go
   
3. Load at runtime
   plugin.Open("loaders/index.so")
   sym, _ := plugin.Lookup("LoadIndex")
   
4. Call function
   loader := sym.(func(*gin.Context) (any, error))
```

#### Problems (Fatal Issues)

**1. Platform Support (BLOCKING)**
```
Supported:  Linux, macOS
NOT Supported:  Windows, WebAssembly, iOS, Android, etc.

HTGO ships as single binary for all platforms
Plugins break this model
```

**2. Deployment Complexity**
```
CURRENT: ./dist/app  (single binary)

WITH PLUGINS:
./dist/app
./dist/loaders/index.so
./dist/loaders/about.so
./dist/loaders/blog.so
... (many .so files)
```

**3. ABI Instability**
```go
// Breaking scenario:
// Built with Go 1.22
plugin.so

// User updates to Go 1.23
go version  // Now 1.23
./app       // Tries to load plugin.so
            // Runtime panic: ABI mismatch!
```

**4. No Error Isolation**
```go
// Plugin has a bug
plugin.Open("bad.so")
sym, _ := plugin.Lookup("BadLoader")
// PANIC: crashes entire app, no graceful recovery
```

**5. Build Complexity**
```bash
# No automatic build-all-plugins support in go build
# Must manually build each plugin separately
$ go build -buildmode=plugin -o loaders/index.so loaders/index.go
$ go build -buildmode=plugin -o loaders/about.so loaders/about.go
$ go build -buildmode=plugin -o loaders/blog.so loaders/blog.go
# Then build main app
$ go build -o dist/app
```

#### Verdict: NOT VIABLE

**Reason**: Breaks HTGO's core design goal of single-binary deployment. Platform limitations alone disqualify this approach.

---

### Approach 2: Reflection-Based Discovery

#### What It Is
Scan a type/package at runtime using Go's `reflect` package to find functions matching the loader signature.

#### How It Would Work

**Option A: Scan Methods on a Type**

```go
// Would need to refactor pages as a struct with methods
type Loaders struct{}

func (l *Loaders) Index(c *gin.Context) (any, error) {
	// Discoverable via reflection
	return nil, nil
}

func (l *Loaders) About(c *gin.Context) (any, error) {
	// Discoverable via reflection
	return nil, nil
}

// Discovery
t := reflect.TypeOf(&Loaders{})
for i := 0; i < t.NumMethod(); i++ {
	method := t.Method(i)
	if hasValidSignature(method) {
		// Register as loader
	}
}
```

**Option B: Scan Package Functions**

```go
// Cannot work - reflection doesn't expose package-level functions
// Only works on types and their methods
func LoadIndex() { }  // ❌ Not reflectable
```

#### Problems

**1. Requires Architectural Change**
```
CURRENT:
pages/index.go:  func LoadIndex(c *gin.Context) (any, error) { }
pages/about.go:  func LoadAbout(c *gin.Context) (any, error) { }

WITH REFLECTION:
pages/loaders.go: 
  type Loaders struct{}
  func (l *Loaders) Index(c *gin.Context) (any, error) { }
  func (l *Loaders) About(c *gin.Context) (any, error) { }
  
This contradicts HTGO's colocated pattern!
```

**2. Loses Compile-Time Safety**
```go
// Typo in function name
func (l *Loaders) Indx(c *gin.Context) (any, error) {  // Typo!
	return nil, nil
}

// No compile error - caught at runtime only!
// User discovers issue when testing

// vs Code Gen approach:
// Compiler immediately detects mismatch between file and function
```

**3. Slow Performance**
```go
// Runtime reflection overhead (1-2 orders of magnitude slower than direct calls)
// Startup time: 100ms → 50-100ms extra
// Not huge, but unnecessary cost for initialization

// vs Code Gen:
// No runtime cost - all work done at build time
```

**4. No IDE Support for Bare Functions**
```go
// IDE can't help with function discovery
// No autocomplete for function names
// Must manually document the convention

// vs Manual/CodeGen:
// IDE shows all available functions
// Autocomplete works
// Cross-reference jumps work
```

**5. Runtime Crashes Possible**
```go
// Reflection can panic on unexpected types
_ = method.Func.Interface().(func(*gin.Context) (any, error))  // Can panic!

// vs Code Gen:
// Compiler ensures type safety
// No runtime panics
```

#### Verdict: Technically Possible But Problematic

**Reasons**:
- Requires architectural refactoring
- Loses compile-time safety
- Adds unnecessary runtime overhead
- Contradicts HTGO's design principles

---

### Approach 3: Build-Time Code Generation (RECOMMENDED)

#### What It Is
Use `go:generate` directive to create a `.go` file that registers all discovered loaders at build time.

#### How It Works

**Step-by-Step**

```
1. Developer writes loaders in pages/*.go
   
2. Run: go generate ./...
   
3. Generator tool:
   - Scans pages/ directory
   - Finds all exported functions matching signature
   - Derives route from file path
   - Generates: pages/loaders_generated.go
   
4. Generated file:
   var LoaderRegistry = map[string]func(...){
       "/":      LoadIndex,
       "/about": LoadAbout,
   }
   
5. Application uses:
   Loaders: pages.LoaderRegistry
```

#### Implementation Example

**Generator Tool** (100 lines of Go)

```go
// cmd/htgo-gen-loaders/main.go
func main() {
    loaders := discoverLoaders(pagesDir)
    generated := generateRegistry(loaders)
    os.WriteFile(outputFile, generated, 0644)
}

func discoverLoaders(pagesDir string) []LoaderInfo {
    // Walk pages directory
    // Parse each .go file with go/ast
    // Find exported functions matching signature
    // Derive route from file path
    // Return []LoaderInfo
}

func generateRegistry(loaders []LoaderInfo) []byte {
    // Build Go code string
    // Write to buffer
    // Return as bytes
}
```

**Integration with CLI**

```bash
$ htgo dev
→ Auto-runs: go generate ./...
→ Generates: pages/loaders_generated.go
→ Starts dev server

$ htgo build
→ Auto-runs: go generate ./...
→ Generates: pages/loaders_generated.go
→ Builds binary
```

#### Advantages

**1. Zero Runtime Overhead**
```
All work happens at build time
No reflection, no dynamic lookups
Generated code is static Go
Identical performance to manual registration
```

**2. Type Safety**
```go
// Generated code is checked by Go compiler
var LoaderRegistry = map[string]func(c *gin.Context) (any, error){
    "/": LoadIndex,  // ← Compiler verifies LoadIndex exists and has right type
}

// If LoadIndex is deleted:
// $ go build
// undefined: pages.LoadIndex  ← Compile error!
```

**3. Single Binary Deployment**
```
Works with go:embed
No external files needed
Perfect for HTGO's model
```

**4. Debuggable**
```go
// Can inspect generated code
$ cat pages/loaders_generated.go

var LoaderRegistry = map[string]func(c *gin.Context) (any, error){
    "/":           LoadIndex,
    "/about":      LoadAbout,
    "/blog/:slug": LoadBlogSlug,
}

// Human-readable, easy to verify
```

**5. IDE Friendly**
```
- Function names in IDE autocomplete
- Jump-to-definition works
- Rename refactoring works
- No IDE plugins needed
```

**6. Aligns with HTGO Philosophy**
```
✓ Boring - just generates Go code
✓ Simple - no runtime magic
✓ Explicit - generated code is visible
✓ Reversible - easy to modify or disable
✓ Maintainable - straightforward logic
```

#### Trade-offs

**Gains**
- Zero boilerplate (no manual listing)
- Auto-discovery (new loaders work immediately)
- Convention-driven (follows file path pattern)

**Losses**
- Extra build step (though auto-integrated)
- Must follow naming convention
- Harder to selectively disable loaders

**Mitigation**
- Auto-run in dev/build commands
- Clear, simple naming rules
- Support manual overrides

#### Verdict: OPTIMAL SOLUTION

**Reasons**:
- Maintains HTGO's simplicity philosophy
- Zero runtime cost
- Type-safe and compiler-checked
- Works on all platforms
- Single-binary deployment
- Reuses existing AST code in router.go

---

### Approach 4: Manual Registration (Current)

#### What It Is
Developer manually lists loaders in a map in `app.go`.

#### How It Works

```go
// examples/sink/app.go
Loaders: map[string]func(c *gin.Context) (any, error){
    "/":      pages.LoadIndex,
    "/about": pages.LoadAbout,
}
```

#### Advantages

**1. Explicit and Flexible**
```
Can map any function to any route
No conventions to follow
Can selectively include/exclude loaders
```

**2. Type Safe**
```go
// Compile error if LoadIndex doesn't exist
"/": pages.LoadIndex,  // ← Compiler checks this
```

**3. Simple**
```
No tool needed
No build step
Just a Go map
```

#### Disadvantages

**1. Boilerplate**
```
Every new loader needs manual entry
Easy to forget
Creates duplication (function is both defined and listed)
```

**2. Maintenance Burden**
```go
// If you add pages/profile.go with LoadProfile()
// You must remember to add to Loaders map
// If you forget, it silently doesn't work
// No error, just confusing behavior
```

**3. Scales Poorly**
```
10 pages: manageable
100 pages: tedious
1000 pages: error-prone
```

#### Verdict: Works, But Not Optimal

**Reasons**:
- Creates boilerplate
- Easy to make mistakes
- Doesn't scale
- Code generation is clearly better

---

## Decision Tree

```
Do you want runtime flexibility?
├─ YES → Go Plugins (if platform support not a concern)
│        But: Breaks HTGO's design
│
└─ NO (preferred for HTGO)
   │
   Can you live with naming convention?
   ├─ YES → Code Generation (RECOMMENDED)
   │        All benefits, minimal trade-offs
   │
   └─ NO → Manual Registration (Current)
            Full flexibility, more boilerplate
```

---

## Recommendation Summary

| Aspect | Recommendation |
|--------|-----------------|
| **Best Overall** | Build-Time Code Generation |
| **Best for Flexibility** | Manual Registration |
| **Best if Platform Agnostic** | Could use Reflection (but HTGO isn't) |
| **Avoid** | Go Plugins (incompatible with HTGO) |

### Why Code Generation Wins for HTGO

1. **Philosophy Match**: Boring, simple, explicit - exactly what HTGO stands for
2. **Zero Cost**: No runtime overhead or complexity
3. **Type Safety**: Compiler guarantees correctness
4. **Developer Experience**: Less boilerplate, auto-discovery works
5. **Platform Support**: Works everywhere (unlike plugins)
6. **Reversibility**: Easy to disable, modify, or go back to manual
7. **Leverage Existing Code**: Can reuse AST parsing from router.go

---

## Next Steps

If implementing code generation:

1. Extract AST utilities from router.go into reusable module
2. Build generator tool in cmd/htgo-gen-loaders/
3. Integrate with dev/build CLI commands
4. Add go:generate directive to examples
5. Document naming convention in CLAUDE.md
6. Update examples to use LoaderRegistry

See `IMPLEMENTATION_GUIDE.md` for detailed implementation steps.

