# Alloy - Go + React SSR Framework

Minimal Go library for server-side rendering (SSR) React applications.

## Overview

**Core concept**: React components → esbuild (dual bundling) → quickjs-go (SSR) + hydration

Features:
- Go backend with Gin web framework
- React for UI (server + client)
- Automatic bundling with esbuild
- Tailwind CSS integration
- Single binary deployment with embedded assets
- Hot-reload development server

## Architecture

### Key Components

| File/Dir | Purpose |
|----------|---------|
| `engine.go` | Core API: routes, rendering setup |
| `router.go` | Route discovery from `pages/` |
| `render.go` | SSR execution via quickjs-go |
| `types.go` | Core types: `Engine`, `Page`, `Options` |
| `bundles.go` | In-memory bundle caching |
| `errors.go` | Error handling with hints |
| `cli/bundle.go` | esbuild dual-bundle generation |
| `cli/dev.go` | Dev mode orchestration |
| `cli/build.go` | Production build pipeline |
| `cli/go-watcher.go` | Go file watcher + restart |
| `cli/pages-watcher.go` | Pages directory monitoring |
| `cli/hot-reload.go` | WebSocket hot reload |
| `cli/tailwind.go` | Tailwind auto-download + plugin |
| `loaderutil/` | AST parsing for loaders & API handlers |

### Page Lifecycle

1. Route → Gin handler
2. Load `.ssr.js` bundle (from cache or disk)
3. Execute in quickjs-go
4. React renders to HTML string
5. Inject props + client bundle URLs
6. Browser loads client bundle → `ReactDOM.hydrateRoot()`

## Quick Start

### New Project
```bash
alloy new my-app
cd my-app
alloy install
alloy dev
```

### Dev Commands
```bash
alloy dev                # Port 8080
alloy dev --port 3000   # Custom port
```

### Production
```bash
alloy build             # Creates dist/app
./dist/app              # Run binary
```

## Configuration

### Basic Setup

In `app.go`:
```go
var Options = alloy.Options{
    Router: gin.Default(),
    Title:  "My App",
    Port:   "8080",  // optional
    Pages: []alloy.Page{
        {
            Route: "/",
            File:  "pages/index.tsx",
        },
    },
}
```

### Page Structure
```
pages/
├── index.tsx        # Component
├── index.go         # Loader (optional)
├── about.tsx
└── blog/
    ├── [slug].tsx
    └── [slug].go
```

### Dynamic Props (Go Loader)

```go
// pages/blog/[slug].go
func LoadBlogPost(c *gin.Context) (any, error) {
    slug := c.Param("slug")
    return map[string]any{
        "title": "Post: " + slug,
    }, nil
}
```

### Component Using Props

```tsx
// pages/blog/[slug].tsx
export default function BlogPost() {
  const props = window.__Alloy_PROPS__ || {};
  return <h1>{props.title}</h1>;
}
```

### Tailwind CSS

```tsx
// pages/index.tsx
import 'tailwind.css';

export default function Home() {
  return <div className="flex items-center justify-center">
    <h1>Hello Alloy</h1>
  </div>;
}
```

## Page Configuration

| Field | Type | Purpose |
|-------|------|---------|
| `Route` | string | URL path (e.g., "/about", "/blog/:id") |
| `File` | string | Path to `.tsx` component |
| `Handler` | func | Dynamic props: `func(*gin.Context)(any, error)` |
| `Props` | any | Static props |
| `Title` | string | Page title |
| `MetaTags` | []MetaTag | SEO metadata |
| `Links` | []Link | Head links |

## Development

### Dev Mode Features

- **Go Watcher** (`cli/go-watcher.go`): Restarts on `.go` changes
- **Pages Watcher** (`cli/pages-watcher.go`): Monitors `.tsx`/`.css`
- **Hot Reload** (`cli/hot-reload.go`): WebSocket browser refresh
- **Bundle Caching** (`bundles.go`): 1-5ms per-request improvement

### Build Output

```
.alloy/
├── pages/
│   ├── index.ssr.js    # Server bundle
│   ├── index.js        # Client bundle
│   └── index.css       # Styles
└── ...
```

## Production Build

```bash
alloy build
# Validates pages → Bundles (parallel) → Minifies → Embeds → dist/app
```

Environment:
- `Alloy_ENV=production`: Minified bundles
- `GIN_MODE=release`: Gin release mode

## Design Principles

- **Simplicity First**: Minimal API, straightforward implementations
- **Zero Config**: Works out of the box
- **Single Binary**: All assets embedded
- **Unopinionated**: Flexible for different use cases

## API Reference

### Core Types

```go
// Engine - manages routing and rendering
type Engine struct {
    Router   *gin.Engine
    Pages    []Page
    Loaders  map[string]interface{}
    Handlers map[string]gin.HandlerFunc
}

// Page - defines a route
type Page struct {
    Route       string
    File        string
    Handler     func(*gin.Context) (any, error)
    Props       any
    Title       string
    MetaTags    []MetaTag
    Links       []Link
}

// Options - engine configuration
type Options struct {
    Router       *gin.Engine
    Title        string
    Port         string
    Pages        []Page
    Loaders      map[string]interface{}
    Handlers     map[string]gin.HandlerFunc
    ErrorHandler func(*gin.Context, error, *Page)
}
```

### Main Functions

```go
// New creates engine
engine := alloy.New(options)

// Register routes to router
engine.RegisterRoutes()

// Register bundle static handler
engine.RegisterBundles()

// Start HTTP server
engine.Listen()

// Or all three at once (deprecated)
engine.Start()
```

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Hot reload not working | Restart: `Ctrl+C` → `alloy dev` |
| Props not appearing | Check handler in server logs; verify JSON serializable |
| Tailwind not applying | Add `import 'tailwind.css'` to component |
| Build fails | Check `.tsx` syntax; verify files exist in `pages/` |
| SSR errors | Check browser console; ensure no browser APIs in component |

## Deployment

### Docker
```dockerfile
FROM golang:1.23
WORKDIR /app
COPY . .
RUN alloy install && alloy build

FROM alpine:latest
WORKDIR /app
COPY --from=0 /app/dist/app .
CMD ["./app"]
```

## Code Style

- Boolean conditions: extract to named variables, never inline
- Minimize dependencies (core: gin, quickjs-go, esbuild, fsnotify)
- Keep it simple: page definitions are just data structs
- No manual webpack/vite config needed

## Recent Cleanup

Removed redundant code:
- Deleted dead generator tools (`cmd/alloy-gen-*`)
- Merged `apiutil` into `loaderutil` (duplicate code)
- Consolidated `page.go` into `engine.go` (only 2 functions)
- Removed empty boilerplate files

**Status**: Production Ready ✓

Last Updated: October 2025
