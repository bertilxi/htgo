# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**HTGO** is a Go library for server-side rendering (SSR) of React applications. It combines:
- Go backend with Gin web framework
- React for UI (both server and client-side)
- TypeScript/TSX support out of the box
- Automatic bundling with esbuild
- Tailwind CSS integration

The framework automatically handles splitting React components into server bundles (executed via quickjs-go for SSR) and client bundles (for hydration), making full-stack React+Go development simple and fast.

## Architecture Overview

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

1. **Engine** (`engine.go`, `types.go`): Core API for defining pages and routes
   - `Options`: Configuration (Router, Pages, global metadata)
   - `Page`: Individual page definition with route, file, props, metadata
   - `New()`: Creates engine instance and sets up Gin routing

2. **Renderer** (`page.go`): Handles page rendering
   - Loads `.ssr.js` bundle
   - Executes in quickjs-go runtime
   - Renders React to HTML string
   - Injects client bundle paths and props
   - Returns final HTML response

3. **Bundler** (`cli/bundle.go`): esbuild integration
   - Creates dual bundles from single `.tsx` file
   - Server bundle: React renderToString wrapper, no browser APIs
   - Client bundle: ReactDOM.hydrateRoot wrapper, browser APIs enabled
   - Automatically applies Tailwind CSS transformations

4. **Developer Tools** (`cli/dev.go`, `cli/hot-reload.go`, `cli/go-watcher.go`):
   - Go file watcher monitors `.go` and `.tsx` files and triggers hot reload
   - Component bundler monitors `.tsx` and `.css` files via esbuild
   - WebSocket-based hot reload system for browser refresh
   - Graceful process restart via `syscall.Exec()` on Go changes

### Page Lifecycle

1. User navigates to a route (e.g., `/page-path`)
2. Gin routes to `page.render()` handler
3. Engine loads corresponding `.ssr.js` bundle
4. quickjs-go executes the server bundle
5. React renders component to HTML string
6. Server injects props and client bundle URLs
7. HTML sent to browser with embedded script/link tags
8. Client bundle loads and calls `ReactDOM.hydrateRoot()`
9. React takes over interactivity client-side

### Directory Structure

| Location | Purpose |
|----------|---------|
| `/` | Core library (engine, types, utils, page rendering) |
| `/cli/` | Build and development tools (bundling, dev server, hot reload) |
| `/examples/minimal/` | Minimal starter example with single page |
| `/examples/sink/` | Complex example with multiple pages and props handlers |
| `.htgo/` | Build cache (git-ignored, created at runtime) |

## Common Development Tasks

### Setup

```bash
# For working with the library itself:
cd /home/berti/Code/3lines/htgo
go mod tidy

# For working on examples:
cd examples/minimal  # or examples/sink
htgo install
```

### Running Examples

```bash
cd examples/minimal

# Development with hot reload
htgo dev

# Build for production
htgo build

# Run production binary
./dist/app
```

### Key Files for Different Tasks

| Task | Files to Check |
|------|------------------|
| Add a new page | `Page` type in `types.go`, examples in `examples/` |
| Configure routing | `engine.go`, particularly `New()` and `HandleRoutes()` |
| Customize bundling | `cli/bundle.go` for esbuild configuration |
| Debug rendering issues | `page.go` for SSR logic, quickjs-go integration |
| Add page metadata | `Page.Title`, `Page.MetaTags`, `Page.Links` in `types.go` |
| Handle dynamic props | `Page.Handler` pattern in `types.go` and examples |
| Modify Tailwind setup | `cli/tailwind.go` and esbuild plugin configuration |

## Key Implementation Patterns

### Page Definition

```go
Page{
    Route:       "/",                    // URL route
    File:        "pages/index.tsx",      // Component file
    Interactive: true,                   // Enable client hydration
    Props:       initialProps,           // Server-side props
    Handler:     dynamicPropsFunc,       // Optional: dynamic props per request
    Title:       "Page Title",
    MetaTags:    []MetaTag{...},
}
```

### Dynamic Props Handler

```go
Handler: func(c *gin.Context) Page {
    // Access gin context for query params, headers, etc.
    id := c.Query("id")
    return Page{
        Props: map[string]interface{}{
            "id": id,
        },
    }
}
```

### Component Communication

- **Server → Client**: Props passed via `window.__HTGO_PROPS__` in rendered HTML
- **Client → Server**: Standard HTTP requests/APIs (page is just React hydrated)
- **CSS**: Import normally, `@import "tailwindcss"` triggers Tailwind processing

## Build System Details

### Development Mode

- **Go file watcher** (`cli/go-watcher.go`): Watches `.go` files in `.`, `cmd/`, `app/`, `pages/` directories
  - On change: rebuilds dev binary to `tmp/bin/dev`
  - Uses `syscall.Exec()` to gracefully replace current process with new binary
  - Debounced to prevent rapid rebuilds (100ms)
- **Component bundler** (`cli/bundle.go`): Watches `.tsx` and `.css` files
  - Rebuilds `.ssr.js` (server bundle) and `.js`/`.css` (client bundles) via esbuild
  - Output written to `.htgo/` subdirectories
- **Hot reload watcher** (`cli/hot-reload.go`): Watches `.htgo/` output directory
  - Detects bundle changes and broadcasts WebSocket "reload" message
  - Browser receives message and calls `window.location.reload()`
  - Uses `htgo.ClearBundleCache()` to ensure fresh bundles are loaded

### Production Build

1. `cli.Build()`: Bundles all pages (sets `HTGO_ENV=production`)
   - Creates `.ssr.js`, `.js`, `.css` in `.htgo/` subdirectories
   - Minifies all output (identifiers, syntax, whitespace)
2. `go build` with `//go:embed .htgo`: Embeds bundles into binary
3. Produces single executable with zero external asset dependencies

### Environment Variables

- `HTGO_ENV=production`: Production mode (minified bundles, no dev features)
- `GIN_MODE=release`: Gin release mode (no request logging)
- Default (unset): Development mode with hot reload

## Code Style & Patterns

### Keep It Simple

- HTGO prioritizes minimal API surface over feature completeness
- Page definitions are just data structs
- Bundling is handled automatically—no manual webpack/vite config needed
- Single-binary deployment is the default

### No Tests Currently

This library does not include unit or integration tests. Examples (`minimal`, `sink`) serve as functional validation. When adding features:
- Manually test in examples
- Verify bundling works correctly
- Ensure hot reload functions as expected

### Dependencies

- **Core**: `gin`, `quickjs-go`, `esbuild`, `fsnotify`, `gorilla/websocket`
- Minimal dependencies by design—avoid adding unnecessary packages
- Built-in hot reload via `cli.Dev()` eliminates need for external tools like `air`

## Troubleshooting Common Issues

| Issue | Solution |
|-------|----------|
| Hot reload not working | Check WebSocket connection; ensure `.htgo/` directory exists |
| Props not appearing | Verify props are serializable to JSON; check `htmlTemplateData` in `page.go` |
| Tailwind CSS not applying | Ensure `@import "tailwindcss"` in CSS or check `cli/tailwind.go` plugin |
| SSR errors | Check quickjs-go console output for JavaScript runtime errors |
| Build artifacts missing | Run `htgo build` first; verify `.htgo/` directory permissions |

## Quick Reference

| Task | Command |
|------|---------|
| Install dependencies | `htgo install` |
| Start dev server | `htgo dev` |
| Build production binary | `htgo build` |
| Run production app | `./dist/app` |
| Check for Go issues | `go vet ./...` |
| Format code | `go fmt ./...` |
