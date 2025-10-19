# CLAUDE.md - Alloy Project Guide

Comprehensive guidance for working with Alloy. This document consolidates all essential information for development.

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Getting Started](#getting-started)
4. [CLI Commands](#cli-commands)
5. [Configuration](#configuration)
6. [Implementation Patterns](#implementation-patterns)
7. [Build System](#build-system)
8. [Recent Improvements (Phases 1-3)](#recent-improvements)
9. [Troubleshooting](#troubleshooting)
10. [Quick Reference](#quick-reference)

---

## Project Overview

**Alloy** is a Go library for server-side rendering (SSR) of React applications. It combines:
- Go backend with Gin web framework
- React for UI (both server and client-side)
- TypeScript/TSX support out of the box
- Automatic bundling with esbuild
- Tailwind CSS integration

The framework automatically handles splitting React components into server bundles (executed via quickjs-go for SSR) and client bundles (for hydration), making full-stack React+Go development simple and fast.

### Philosophy
- **Simplicity First**: Minimal API surface, straightforward implementations
- **Single Binary Deployment**: All assets embedded, no external files
- **Zero Configuration**: Works out of the box
- **Unopinionated Design**: Flexible for different use cases

---

## Architecture

### Core Flow

```
React Components (.tsx)
        ↓
    [esbuild]
    ↙         ↘
Server SSR    Client Bundle
Bundle        (.js + .css)
(.ssr.js)
    ↓         ↓
 quickjs-go  Browser
    ↓         ↑
HTML String  Hydrate
    ↓         ↑
  HTML + Client Bundle Injection
```

### Key Components

1. **Engine** (`engine.go`, `types.go`): Core API
   - `Options`: Configuration (Router, Pages, global metadata)
   - `Page`: Individual page definition with route, file, props, metadata
   - `New()`: Creates engine instance and sets up Gin routing

2. **Renderer** (`render.go`): Handles page rendering
   - Loads `.ssr.js` bundle
   - Executes in quickjs-go runtime
   - Renders React to HTML string
   - Injects client bundle paths and props
   - Enhanced error handling with helpful hints

3. **Bundler** (`cli/bundle.go`): esbuild integration
   - Creates dual bundles from single `.tsx` file
   - Server bundle: React renderToString wrapper, no browser APIs
   - Client bundle: ReactDOM.hydrateRoot wrapper, browser APIs enabled
   - Automatically applies Tailwind CSS transformations
   - Parallel builds for improved speed (2-3x faster)

4. **Developer Tools** (`cli/dev.go`, `cli/hot-reload.go`, `cli/go-watcher.go`):
   - Go file watcher monitors `.go` and `.tsx` files
   - Component bundler monitors `.tsx` and `.css` files via esbuild
   - WebSocket-based hot reload system with multi-client support
   - Graceful process restart via `syscall.Exec()` on Go changes

5. **Bundle Caching** (`bundles.go`): Performance optimization
   - In-memory caching of `.ssr.js` bundles
   - 1-5ms per-request improvement
   - Automatic cache invalidation on rebuild

### Page Lifecycle

1. User navigates to a route (e.g., `/page-path`)
2. Gin routes to page handler
3. Engine loads corresponding `.ssr.js` bundle (from cache or disk)
4. quickjs-go executes the server bundle
5. React renders component to HTML string
6. Server injects props and client bundle URLs
7. HTML sent to browser with embedded script/link tags
8. Client bundle loads and calls `ReactDOM.hydrateRoot()`
9. React takes over interactivity client-side

### Directory Structure

| Location | Purpose |
|----------|---------|
| `/` | Core library (engine, types, rendering, bundling) |
| `/cli/` | Build and development tools (bundling, dev server, hot reload) |
| `/examples/minimal/` | Minimal starter example with single page |
| `/examples/sink/` | Complex example with multiple pages and props handlers |
| `.alloy/` | Build cache (git-ignored, created at runtime) |

---

## Getting Started

### Setup for Library Development

```bash
cd /home/berti/Code/3lines/alloy
go mod tidy
```

### Setup for Example Projects

```bash
cd examples/minimal  # or examples/sink
alloy install
```

### Create New Project

```bash
alloy new my-app
cd my-app
alloy install
alloy dev
```

---

## CLI Commands

### `alloy new <name>`

Creates a complete Alloy project with:
- Page structure (`pages/` directory)
- Command entry points (`cmd/dev`, `cmd/build`, `cmd/app`)
- Example welcome page with Tailwind styling
- Proper configuration in `app.go`

**Usage:**
```bash
alloy new my-app
cd my-app
alloy install
alloy dev
```

### `alloy dev [options]`

Starts development server with hot-reload.

**Options:**
- `--port <number>` - Custom port (default: 8080)
- `--dir <path>` - Project directory (default: current)

**What it does:**
1. Shows startup banner with routes list
2. Starts dev server with file watchers
3. Enables hot-reload on file changes
4. Multi-client WebSocket support for browser tabs

**Example:**
```bash
alloy dev              # Port 8080
alloy dev --port 3000 # Custom port
```

### `alloy build [options]`

Builds production-ready binary with embedded assets.

**Options:**
- `--dir <path>` - Project directory (default: current)
- `--output <path>` - Output binary location

**What it does:**
1. Pre-validates all page files exist
2. Shows validation errors or starts bundling
3. Displays progress for each page
4. Outputs binary to `dist/app`

**Example:**
```bash
alloy build
alloy build --dir ./myapp
```

### `alloy --help`

Shows all available commands and options.

### `alloy version`

Shows CLI version information.

---

## Configuration

### Page Structure

Create pages in `pages/` directory. Each page needs:
- **Component file** (`.tsx`): React component
- **Optional loader file** (`.go`): Dynamic props handler

**Example Directory:**
```
pages/
├── index.tsx          # Home page component
├── index.go          # Home page loader (optional)
├── about.tsx         # About page component
└── blog/
    ├── [slug].tsx    # Blog post component
    └── [slug].go     # Blog loader with slug param
```

### App Configuration

In `app.go`:

```go
var Options = alloy.Options{
    Router: gin.Default(),
    Title:  "My Alloy App",
    Port:   "8080",              // Optional - defaults to 8080
    Pages: []alloy.Page{
        {
            Route:       "/",
            File:        "pages/index.tsx",
            Interactive: true,              // Enable hydration
            Title:       "Home",
            MetaTags: []alloy.MetaTag{
                {Name: "description", Content: "Home page"},
            },
        },
        {
            Route:   "/about",
            File:    "pages/about.tsx",
            Handler: LoadAbout,             // Optional - dynamic props
        },
    },
}
```

### Page Options

| Field | Type | Purpose |
|---|---|---|
| Route | string | URL path (e.g., "/about", "/blog/:id") |
| File | string | Page component file |
| Interactive | bool | Enable client hydration |
| Props | any | Static data passed to component |
| Handler | func | Dynamic props per request |
| Title | string | Page title |
| MetaTags | []MetaTag | SEO metadata |
| Links | []Link | Head links |

### Error Handler (Optional)

Handle rendering errors with custom logic:

```go
Options{
    ErrorHandler: func(c *gin.Context, err error, page *Page) {
        // Log error, render custom error page, etc.
        c.JSON(500, gin.H{"error": err.Error()})
    },
}
```

### Asset Configuration (Optional)

Customize asset URLs and cache busting:

```go
Options{
    AssetURLPrefix:   "https://cdn.example.com/assets",
    CacheBustVersion: "v1.2.0",
}
```

---

## Implementation Patterns

### Basic Page Component

```tsx
// pages/index.tsx
export default function Home() {
  return <div className="flex items-center justify-center">
    <h1>Welcome to Alloy</h1>
  </div>;
}
```

### Page with Interactivity

```tsx
// pages/counter.tsx
import { useState } from 'react';

export default function Counter() {
  const [count, setCount] = useState(0);

  return (
    <div>
      <p>Count: {count}</p>
      <button onClick={() => setCount(count + 1)}>
        Increment
      </button>
    </div>
  );
}
```

### Dynamic Props Handler

```go
// pages/blog.go
package pages

import (
    "github.com/gin-gonic/gin"
)

func LoadBlog(c *gin.Context) (any, error) {
    slug := c.Param("slug")
    if slug == "" {
        return nil, errors.New("slug required")
    }

    return map[string]any{
        "slug": slug,
        "title": "Blog Post: " + slug,
    }, nil
}
```

### Component Using Props

```tsx
// pages/blog/[slug].tsx
declare global {
  interface Window {
    __Alloy_PROPS__: any;
  }
}

export default function BlogPost() {
  const props = window.__Alloy_PROPS__ || {};

  return (
    <article>
      <h1>{props.title}</h1>
      <p>Slug: {props.slug}</p>
    </article>
  );
}
```

### Tailwind CSS

```tsx
// pages/index.tsx
import 'tailwind.css';

export default function Home() {
  return (
    <div className="flex items-center justify-center h-screen bg-gradient-to-r from-blue-500 to-purple-600">
      <h1 className="text-4xl text-white">Hello Alloy!</h1>
    </div>
  );
}
```

### Component Communication

- **Server → Client**: Props passed via `window.__Alloy_PROPS__`
- **Client → Server**: Standard HTTP requests (fetch, axios, etc.)
- **CSS**: Import normally, `@import "tailwindcss"` enables Tailwind

---

## Build System

### Development Mode

**Go Watcher** (`cli/go-watcher.go`):
- Watches `.go` files in `.`, `cmd/`, `app/`, `pages/` directories
- Rebuilds and restarts dev binary on change
- Uses `syscall.Exec()` for graceful restart
- Debounced (100ms) to prevent rapid rebuilds

**Component Bundler** (`cli/bundle.go`):
- Watches `.tsx` and `.css` files
- Rebuilds bundles via esbuild
- Output to `.alloy/` subdirectories
- Parallel builds for speed

**Hot Reload** (`cli/hot-reload.go`):
- Watches `.alloy/` output directory
- Broadcasts WebSocket "reload" messages
- Browser auto-refreshes on change
- Multi-client support for multiple tabs

### Production Build

1. **Pre-validation**: Checks all page files exist and are valid
2. **Parallel Bundling**: Builds all pages concurrently
3. **Minification**: All output minified (identifiers, syntax, whitespace)
4. **Embedding**: Bundles embedded into binary via `//go:embed .alloy`
5. **Output**: Single executable in `dist/app`

**Environment Variables:**
- `Alloy_ENV=production`: Minified bundles, no dev features
- `GIN_MODE=release`: Gin release mode (no logging)

### Build Output Structure

```
.alloy/
├── pages/
│   ├── index.ssr.js      # Server-side rendering bundle
│   ├── index.js          # Client-side hydration bundle
│   ├── index.css         # Component styles
│   ├── about.ssr.js
│   ├── about.js
│   └── about.css
└── ...
```

### Performance Improvements (Phase 2)

- **Bundle Caching**: 1-5ms saved per request (in-memory cache)
- **Parallel Builds**: 2-3x faster (10 pages: 1-2s → 300-400ms)
- **Concurrent Cache Access**: ~1.7ns overhead (negligible)

---

## Recent Improvements

### Phase 1: Developer Feedback

**Startup Banner**
- Shows server ready status and port
- Lists all registered routes
- Indicates hot-reload is enabled

**Dynamic WebSocket**
- Automatically detects server hostname
- Works on any port, not just 8080
- Works remotely with proper hostname

**Better Error Messages**
- Shows which step failed (props, SSR, bundling, template)
- Provides helpful hints based on error type
- Includes page route and file for debugging

### Phase 2: CLI Tool

**Complete Solution**
- `alloy new` - Project scaffolding
- `alloy dev` - Development server
- `alloy build` - Production builds
- Beautiful, consistent output

**Benefits**
- One-command project creation
- Simple, memorable commands
- Professional developer experience
- Discoverability via `alloy --help`

### Phase 3: Build Process

**Pre-Build Validation**
- Checks all page files exist
- Validates file types (.tsx, .jsx, .ts, .js)
- Detects empty files
- Shows clear validation errors

**Build Progress**
- Shows page count and list
- Per-page progress indicators
- Success/failure status for each page
- Build summary with next steps

**Better Error Context**
- Extracts first esbuild error
- Provides helpful hint based on error type
- Shows route and file for each failure
- Continues bundling other pages

---

## Key Files for Different Tasks

| Task | Files to Check |
|------|------------------|
| Add a new page | `Page` type in `types.go`, examples in `examples/` |
| Configure routing | `engine.go`, particularly `New()` |
| Customize bundling | `cli/bundle.go` for esbuild configuration |
| Debug rendering issues | `render.go` for SSR logic, `errors.go` for error handling |
| Add page metadata | `Page.Title`, `Page.MetaTags`, `Page.Links` in `types.go` |
| Handle dynamic props | `Page.Handler` pattern (see Implementation Patterns section) |
| Modify Tailwind setup | `cli/tailwind.go` and esbuild plugin configuration |
| Add error handling | `types.go` Option field `ErrorHandler` |

---

## Troubleshooting

### Hot Reload Not Working

**Checklist:**
- Dev server running?
- WebSocket connection successful (check browser console)
- `.alloy/` directory exists?
- Correct port in dev command?

**Fix:**
- Restart dev server: `Ctrl+C` → `alloy dev`
- Check browser console for WebSocket errors
- Ensure `.alloy/` directory exists and is writable

### Props Not Appearing

**Checklist:**
- Props serializable to JSON?
- Handler function returns correct type?
- Component accessing `window.__Alloy_PROPS__`?

**Fix:**
- Check handler error in server logs
- Verify props can serialize to JSON (no functions, circular refs)
- Ensure component imports props correctly

### Tailwind CSS Not Applying

**Checklist:**
- CSS file imported?
- Contains `@import "tailwindcss"`?
- Tailwind config exists?

**Fix:**
- Add to component: `import 'tailwind.css'`
- Ensure CSS has `@import "tailwindcss"`
- Check `cli/tailwind.go` configuration

### Build Errors

**Common Issues:**
- Missing npm dependencies: `npm install <package>`
- TypeScript/syntax errors: Fix JSX syntax in component
- Missing page file: Check filename and path in config
- Invalid file extension: Use `.tsx`, not `.ts`

**Debug:**
- Read build output carefully - includes helpful hints
- Check browser console for runtime errors
- Look for "Error at stage" messages in error output

### SSR Errors in Console

**Solutions:**
- Check server logs for JavaScript runtime errors
- Verify component imports are correct
- Ensure no browser-only APIs used in component
- Use `typeof window !== 'undefined'` for browser APIs

---

## Deployment

### Production Build

```bash
alloy build
```

Produces: `dist/app` (single binary with all assets)

### Running Production

```bash
./dist/app              # Port 8080 (default)
PORT=3000 ./dist/app    # Custom port
```

### Docker Example

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

---

## Code Style & Patterns

### Keep It Simple

- Alloy prioritizes minimal API surface
- Page definitions are just data structs
- Bundling is handled automatically
- Single-binary deployment is the default
- No manual webpack/vite config needed

### Boolean Conditions

Always extract to named variables, never inline:

```go
// Good
hasValidData := data && len(data) > 0
if hasValidData { ... }

// Bad
if data && len(data) > 0 { ... }
```

### Dependencies

- **Core**: `gin`, `quickjs-go`, `esbuild`, `fsnotify`, `gorilla/websocket`
- Minimal dependencies by design
- Avoid adding unnecessary packages
- Built-in hot reload eliminates need for external tools

### Testing

No unit tests currently. Examples serve as functional validation. When adding features:
- Manually test in examples
- Verify bundling works correctly
- Ensure hot reload functions as expected

---

## Quick Reference

### Commands

| Task | Command |
|------|---------|
| Create new project | `alloy new my-app` |
| Install dependencies | `alloy install` |
| Start dev server | `alloy dev` |
| Build production | `alloy build` |
| Run production | `./dist/app` |
| Show help | `alloy --help` |
| Show version | `alloy version` |

### Dev Server

| Action | Trigger |
|--------|---------|
| Hot reload | Save file → auto-refresh |
| Restart server | `Ctrl+C` → `alloy dev` |
| Custom port | `alloy dev --port 3000` |

### Go Commands

| Task | Command |
|------|---------|
| Format code | `go fmt ./...` |
| Check errors | `go vet ./...` |
| Tidy modules | `go mod tidy` |
| Run tests | `go test ./...` |

### File Locations

```
project/
├── .alloy/              # Build cache (git-ignored)
├── pages/              # React components
│   ├── index.tsx
│   ├── index.go        # Optional loader
│   └── ...
├── cmd/
│   ├── dev/main.go     # Dev entry point
│   ├── build/main.go   # Build entry point
│   └── app/main.go     # Production entry point
├── app.go              # Project configuration
├── package.json        # npm dependencies
└── .gitignore
```

### Environment

| Variable | Purpose | Default |
|----------|---------|---------|
| `PORT` | Server port | 8080 |
| `Alloy_ENV` | Set to "production" for build | unset (dev) |
| `GIN_MODE` | Set to "release" for production | unset (debug) |

---

## Future Roadmap

### Planned Improvements

**Phase 4: Advanced Tools**
- Configuration validation (`alloy validate`)
- Component scaffolding (`alloy generate`)
- Better error messages for common mistakes
- Bundle size analysis (`alloy analyze`)

**Phase 5: Documentation**
- Expanded README with tutorials
- API reference documentation
- Troubleshooting guides
- Example gallery

**Phase 6: Community**
- Template registry
- Plugin system
- IDE extensions
- Best practices guide

---

## Architecture Decision: Handler Signature

**Current Pattern:**
```go
Handler: func(c *gin.Context) (any, error)
```

**Why This Design:**
- Clear responsibility: return props or error
- Explicit error handling
- No implicit page mutation
- Simpler mental model
- Unopinionated (your handler, your logic)

**Benefits:**
- Handlers can fail gracefully
- Errors are explicit and visible
- Type-safe with compiler checking
- Works with any error handling pattern

---

**Last Updated:** October 2025
**Status:** Production Ready ✓
