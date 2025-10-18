# HTGO Quick Reference

Fast lookup guide for HTGO features, commands, and architecture.

---

## CLI Commands

### `htgo new <name>`
Creates a new HTGO project with complete structure
```bash
htgo new my-app          # Creates my-app/ with all files
cd my-app && htgo install && htgo dev    # Start developing
```

### `htgo dev`
Starts development server with hot-reload
```bash
htgo dev                 # Port 8080
htgo dev --port 3000    # Custom port
```

### `htgo build`
Builds production bundles
```bash
htgo build              # Build in current directory
htgo build --dir ./app  # Build from different directory
```

### `htgo --help`
Shows all available commands and options

### `htgo version`
Shows CLI version

---

## Project Configuration

### app.go Structure
```go
var Options = htgo.Options{
    Router: gin.Default(),
    Title:  "App Title",
    Pages: []htgo.Page{
        {
            Route:       "/",
            File:        "pages/index.tsx",
            Interactive: true,
            Props:       nil,
        },
    },
}
```

### Page Options
| Field | Type | Purpose |
|---|---|---|
| Route | string | URL path (e.g., "/about") |
| File | string | Page component file |
| Interactive | bool | Enable client hydration |
| Props | any | Data passed to component |
| Handler | func | Dynamic props per request |
| Title | string | Page title |
| MetaTags | []MetaTag | SEO metadata |
| Links | []Link | Head links |

---

## Development Workflow

### Create → Dev → Build → Deploy

```
1. Create project:     htgo new my-app
2. Install deps:       htgo install
3. Start dev server:   htgo dev
4. Edit pages:         pages/index.tsx
5. See hot-reload:     Browser auto-refreshes
6. Build production:   htgo build
7. Run production:     htgo start
```

---

## File Structure

```
project/
├── .htgo/              # Build cache (git-ignored)
├── pages/
│   └── index.tsx       # Page components
├── cmd/
│   ├── dev/main.go     # Dev entry point
│   ├── build/main.go   # Build entry point
│   └── app/main.go     # Production entry point
├── app.go              # Project configuration
#
├── package.json        # npm dependencies
└── .gitignore          # Ignore rules
```

---

## Common Tasks

### Create a New Page

1. Create file: `pages/about.tsx`
2. Add component:
```tsx
export default function About() {
  return <div>About Page</div>;
}
```

3. Register in app.go:
```go
{
    Route: "/about",
    File:  "pages/about.tsx",
}
```

### Add Interactivity

Set `Interactive: true` in page config to enable client-side React:
```go
{
    Route:       "/",
    File:        "pages/index.tsx",
    Interactive: true,  // ← Enables hydration
}
```

### Pass Props to Component

```go
{
    Route: "/blog/:id",
    File:  "pages/blog.tsx",
    Handler: func(c *gin.Context) htgo.Page {
        id := c.Param("id")
        return htgo.Page{
            Props: map[string]interface{}{
                "id": id,
            },
        }
    },
}
```

Component receives props via `window.PAGE_PROPS`:
```tsx
export default function Blog() {
  const props = window.PAGE_PROPS;
  return <div>Blog {props.id}</div>;
}
```

### Use Tailwind CSS

Tailwind is automatically processed. Just import:
```tsx
import 'tailwind.css';

export default function Component() {
  return <div className="flex items-center justify-center">
    Hello
  </div>;
}
```

### Add Dependencies

```bash
npm install lodash
npm install --save-dev @types/lodash
```

---

## Build Output

After `htgo build`:
- Binary: `dist/app`
- Bundles embedded: In Go binary
- Size: Single executable (~15MB)

```bash
HTGO_ENV=production GIN_MODE=release ./dist/app
# Server runs on port 8080
```

---

## Error Messages & Solutions

### "File not found: pages/about.tsx"
✓ Check filename spelling
✓ Verify file extension is .tsx
✓ Check file is in pages/ directory

### "Cannot find module 'react'"
✓ Run `npm install react`
✓ Check package.json has "react" in dependencies

### "SyntaxError in component"
✓ Check TypeScript/JSX syntax
✓ Verify imports are correct
✓ Check for missing closing tags

### WebSocket connection failed
✓ Check dev server is running
✓ Verify port matches (should auto-detect)
✓ Check browser console for errors

---

## Performance Tips

- **Code splitting:** Separate pages into different files
- **Dependencies:** Use tree-shakeable libraries (lodash-es)
- **Images:** Optimize size before importing
- **Fonts:** Use system fonts or web-safe fonts
- **CSS:** HTGO handles CSS automatically

---

## Deployment

### Production Build
```bash
htgo build     # Creates dist/app binary
htgo start     # Run in production
```

### Environment Variables
```bash
# Production mode
HTGO_ENV=production
GIN_MODE=release

# Optional: Custom port
PORT=3000
```

### Docker Example
```dockerfile
FROM golang:1.23

WORKDIR /app
COPY . .

RUN htgo build

CMD ["./dist/app"]
```

---

## Documentation Files

| File | Purpose |
|---|---|
| CLAUDE.md | Repository guidance for Claude Code |
| PHASE1_UX_IMPROVEMENTS.md | Dev feedback system |
| PHASE2_CLI_TOOL.md | CLI tool details |
| PHASE3_BUILD_IMPROVEMENTS.md | Build validation |
| UX_IMPROVEMENTS_SUMMARY.md | Complete overview |
| NEXT_STEPS.md | Future improvements roadmap |

---

## Useful Links

- **GitHub:** https://github.com/bertilxi/htgo
- **Examples:** `/examples/minimal`, `/examples/sink`
- **Dependencies:** see go.mod
- **React Docs:** https://react.dev
- **Tailwind CSS:** https://tailwindcss.com

---

## Keyboard Shortcuts (Dev Mode)

| Action | Trigger |
|---|---|
| Hot reload | Save file → auto-refresh |
| Restart server | `Ctrl+C` → `htgo dev` |
| Clear console | `Ctrl+K` (varies by terminal) |

---

## Troubleshooting Checklist

- [ ] Run `htgo install` to install all dependencies
- [ ] Check `app.go` configuration is correct
- [ ] Verify page files exist and are in pages/
- [ ] Check file extensions (.tsx, not .ts)
- [ ] Verify imports are correct
- [ ] Look at browser console for errors
- [ ] Check server logs for error messages
- [ ] Try `htgo clean && htgo install`
- [ ] Restart dev server with `Ctrl+C` + `htgo dev`

---

## Quick Commands

```bash
# Setup
htgo new project
cd project
htgo install

# Development
htgo dev              # Start server
htgo build            # Build for production
htgo start            # Run production binary

# Cleanup
rm -rf dist/          # Remove build artifacts
rm -rf node_modules/  # Remove npm packages
htgo clean            # Full cleanup (if defined)

# Go commands
go mod tidy           # Clean dependencies
go fmt ./...          # Format code
go vet ./...          # Check for errors
```

---

## FAQs

**Q: Can I use custom middleware?**
A: Yes, pass custom Gin router to Options:
```go
router := gin.Default()
router.Use(customMiddleware)
Options.Router = router
```

**Q: Do I need to restart the server after editing app.go?**
A: Yes, changes to app.go require server restart. Edit components (.tsx) for hot-reload.

**Q: Can I use environment variables?**
A: Yes, via os.Getenv() in Go handlers.

**Q: Is SSR necessary?**
A: SSR is on by default. Set `Interactive: false` for static pages.

**Q: How do I debug TypeScript errors?**
A: Check build output. Errors show in `htgo build` output.

**Q: Can I use external CSS frameworks?**
A: Yes, Tailwind is included. Import others via npm.

---

## Version Info

- **CLI Version:** 0.1.0
- **Requires:** Go 1.23.5+, Node.js 18+
- **License:** Check repository

---

**Last Updated:** October 18, 2025
**Status:** Production Ready ✓
