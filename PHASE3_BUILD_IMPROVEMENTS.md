# Phase 3: Build Improvements

## Summary

Enhanced the build process with pre-build validation, progress feedback, and better error messages. Users now see exactly what's happening during builds, errors are caught before bundling, and build failures are clearly explained.

---

## What's Improved

### 1. Pre-Build Validation

Before attempting to bundle pages, the build process now validates:

**File Existence Check**
- All page.File paths must exist
- Pages pointing to non-existent files are caught immediately
- Clear error message showing the route and missing file

**File Type Validation**
- Only `.tsx`, `.jsx`, `.ts`, `.js` files are allowed
- Other file types are rejected with clear message
- Helps catch typos in file paths

**Directory Check**
- Page.File must be a file, not a directory
- Prevents confusion from misconfiguration

**Empty File Detection**
- Warns about pages with empty file content
- Helps catch accidental blank files

**Component Export Detection**
- Warns if page file might not export a default component
- Catches a common mistake early

### Example Validation

```bash
$ htgo build

📦 Starting Production Build
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
❌ Build validation failed:

  1. Route '/about':
     Error: File not found: open pages/about.tsx: no such file or directory
     File: pages/about.tsx

  2. Route '/blog':
     Error: Invalid file extension '.ts'. Expected .tsx, .jsx, .ts, or .js
     File: pages/blog.ts
```

---

### 2. Progress Feedback During Build

Users now see clear progress as each page is bundled:

```bash
📦 Starting Production Build
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📄 Pages to bundle: 3
   • / → pages/index.tsx
   • /about → pages/about.tsx
   • /blog → pages/blog.tsx

📌 Bundling / (pages/index.tsx)...
✓ / bundled
📌 Bundling /about (pages/about.tsx)...
✓ /about bundled
📌 Bundling /blog (pages/blog.tsx)...
✓ /blog bundled

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Production Build Complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📦 Successfully bundled 3 pages

Next steps:
  • Run: go run cmd/app/main.go
  • Or build binary: htgo build && ./dist/app

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**What Each Message Means:**
- `📦 Starting Production Build` - Build has started
- `📄 Pages to bundle: X` - Count and list of all pages
- `📌 Bundling <route>` - Currently working on this page
- `✓ <route> bundled` - Page completed successfully
- `✓ Production Build Complete` - All pages done
- `Next steps` - How to proceed

---

### 3. Better Error Messages

Build errors are now contextualized:

#### Before
```
failed to build client bundle: Cannot find module 'react'
```

#### After
```
❌ /about failed: client bundle error: Module import error: Check npm dependencies
```

**Error Context Hints:**

| Error Type | Hint |
|---|---|
| Cannot find module | Check that imported modules exist and are installed |
| Module not found | Module import error: Check npm dependencies |
| SyntaxError | TypeScript/JSX syntax error: Check component syntax |
| Unexpected token | Parsing error: Invalid syntax in component |
| Invalid JSX | Invalid JSX: Check component JSX syntax |

**Example Build Failure:**

```bash
📌 Bundling /about (pages/about.tsx)...
❌ /about failed: server bundle error: Module import error: Check npm dependencies

📌 Bundling /blog (pages/blog.tsx)...
✓ /blog bundled

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
❌ Build failed: 1 of 3 pages could not be bundled
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### 4. Build Summary Output

After successful builds, clear summary:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Production Build Complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📦 Successfully bundled 5 pages
⚠️  2 warnings

Next steps:
  • Run: go run cmd/app/main.go
  • Or build binary: htgo build && ./dist/app

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Implementation Details

### New File: `cli/buildutils.go`

Contains:
- `ValidatePages()` - Pre-build validation
- `PrintValidationResults()` - Display validation errors
- `PrintBuildStart()` - Opening banner
- `PrintPageBuildStart()` / `PrintPageBuildComplete()` - Progress
- `PrintBuildComplete()` - Success summary
- `PrintBuildFailed()` - Failure summary
- `ExtractBuildErrorContext()` - Error hint generation

### Enhanced File: `cli/build.go`

Now:
1. Prints build start banner
2. Validates all pages before bundling
3. Prints per-page progress
4. Tracks failures
5. Displays completion or failure summary

### Enhanced File: `cli/bundle.go`

- `buildBackend()` - Better error context
- `buildClient()` - Better error context
- Extracts first esbuild error
- Generates helpful hint

---

## Usage Example

### Successful Build

```bash
$ htgo build

📦 Starting Production Build
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📄 Pages to bundle: 2
   • / → pages/index.tsx
   • /about → pages/about.tsx

📌 Bundling / (pages/index.tsx)...
✓ / bundled
📌 Bundling /about (pages/about.tsx)...
✓ /about bundled

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Production Build Complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📦 Successfully bundled 2 pages

Next steps:
  • Run: go run cmd/app/main.go
  • Or build binary: htgo build && ./dist/app

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Failed Validation

```bash
$ htgo build

📦 Starting Production Build
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
❌ Build validation failed:

  1. Route '/about':
     Error: File not found: open pages/about.tsx: no such file or directory
     File: pages/about.tsx

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
❌ Build failed: 1 of 2 pages could not be bundled
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Error Handling Strategy

### Three Levels of Error Checking

**1. Validation Phase** (Before bundling)
- Check files exist
- Check file types
- Check basic configuration
- Stop immediately if validation fails

**2. Bundling Phase** (During build)
- Catch esbuild errors
- Extract first error message
- Provide context hint
- Continue with other pages

**3. Summary Phase** (After bundling)
- Report success or failure
- Show which pages failed
- Provide next steps

### Benefits

1. **Early Detection** - Invalid configs caught before any bundling
2. **Partial Success** - If one page fails, others still build
3. **Clear Feedback** - User sees exactly what failed and why
4. **Actionable Help** - Hints guide toward solutions

---

## Integration with Other Phases

Works seamlessly with previous phases:

**Phase 1 - UX Improvements:**
- Better errors match the enhanced error handling from Phase 1
- Dynamic WebSocket works during builds

**Phase 2 - CLI Tool:**
- `htgo build` command uses these improvements
- Clear feedback matches CLI's polished output

---

## Future Enhancement Ideas

1. **Size Reports**
   - Show bundle sizes per page
   - Compare development vs production sizes
   - Alert on unusually large bundles

2. **Build Optimization**
   - Parallel page bundling (currently sequential)
   - Incremental builds (only changed pages)
   - Cache optimizations

3. **Validation Enhancements**
   - Static import analysis
   - Props shape validation
   - More thorough TypeScript checks

4. **Build Profiling**
   - Time spent per page
   - Bottleneck identification
   - Suggestions for optimization

---

## Files Modified

| File | Changes |
|---|---|
| `cli/buildutils.go` | NEW - Validation and feedback functions |
| `cli/build.go` | Enhanced - Progress tracking and error handling |
| `cli/bundle.go` | Enhanced - Better error context in buildBackend/buildClient |

**Total Changes:**
- New file: ~180 lines
- Modified: ~50 lines
- No breaking changes

---

## Testing the Changes

### Test 1: Valid Build
```bash
cd examples/minimal
htgo build
```
Should show: ✓ All pages bundled successfully

### Test 2: Validation Error
Add invalid route to app.go:
```go
{
    Route: "/broken",
    File: "pages/missing.tsx",
}
```
Then: `htgo build`
Should show: ❌ Validation error for missing file

### Test 3: Build Error
Break component syntax in pages/index.tsx, then: `htgo build`
Should show: Clear error message with hint about fixing syntax

---

## Summary

Phase 3 transforms the build process from silent and opaque to clear and informative:

✅ **Pre-build Validation** - Catch configuration errors early
✅ **Progress Feedback** - Users see what's being built
✅ **Better Error Context** - Hints help fix problems
✅ **Clear Success/Failure** - Professional output
✅ **No Breaking Changes** - Works with existing projects

The build process is now a first-class feature that guides users toward successful builds.
