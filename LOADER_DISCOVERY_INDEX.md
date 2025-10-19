# Loader Auto-Discovery Investigation: Complete Documentation Index

## Overview

This folder contains a comprehensive investigation into automatically discovering and loading Go loader functions in HTGO without manual registration. Four detailed reports explore different technical approaches and recommend the optimal solution.

**Total Documentation**: 1,900+ lines across 4 documents

---

## Documents in This Investigation

### 1. LOADER_AUTODISCOVERY_SUMMARY.md
**Quick Reference** (189 lines)

Start here for a high-level overview.

- Findings at a glance
- Recommended solution: Build-time code generation
- Current HTGO architecture overview
- Advantages and trade-offs
- Risk assessment
- Implementation roadmap

**Best for**: Getting oriented, executive summary, quick decisions

---

### 2. GO_PLUGIN_EXPLORATION.md
**Complete Technical Deep-Dive** (574 lines)

Comprehensive analysis of all approaches and why code generation is recommended.

**Sections:**
1. Go Plugin System - Why it's not viable for HTGO
2. Current HTGO Architecture - What's already in place
3. Reflection-Based Discovery - Technical pros and cons
4. Build-Time Code Generation - Detailed explanation
5. Comparison Matrix - Side-by-side of all approaches
6. Trade-offs Analysis - What we gain/lose with each
7. Alternative: Annotation-Based Registration
8. Recommended Solution with roadmap
9. Naming Convention Details
10. Implementation Checklist
11. Risks and Mitigation
12. Appendix with code examples

**Best for**: Understanding all options, learning the full context, detailed decision-making

---

### 3. APPROACHES_COMPARISON.md
**Side-by-Side Analysis** (503 lines)

Detailed comparison of Go Plugins, Reflection, Code Generation, and Manual approaches.

**Sections:**
- Executive Decision Matrix (all criteria)
- Detailed Analysis of each approach
  - What it is
  - How it works
  - Problems/Advantages
  - Verdict
- Decision Tree
- Recommendation Summary

**Best for**: Comparing specific approaches, understanding trade-offs, convincing stakeholders

---

### 4. IMPLEMENTATION_GUIDE.md
**Technical How-To** (636 lines)

Step-by-step guide to implementing code generation for loader discovery.

**Sections:**
1. Understanding current code (router.go analysis)
2. Code generation tool implementation (complete source code)
3. Integration with go:generate
4. Updated application code (before/after)
5. Example output
6. Integration with htgo CLI
7. Testing the implementation (test cases)
8. Error handling
9. Git integration
10. Migration path
11. Advanced: Manual overrides

**Best for**: Actually implementing the solution, code examples, migration planning

---

## Quick Navigation

### I want to...

**Understand if this is worth pursuing**
→ Read: LOADER_AUTODISCOVERY_SUMMARY.md

**Understand why plugins won't work**
→ Read: GO_PLUGIN_EXPLORATION.md sections 1, 11
→ Or: APPROACHES_COMPARISON.md section "Go Plugins"

**Compare all options objectively**
→ Read: APPROACHES_COMPARISON.md (entire)

**Understand the recommended approach**
→ Read: GO_PLUGIN_EXPLORATION.md section 4
→ Or: LOADER_AUTODISCOVERY_SUMMARY.md "Recommended Solution"

**Understand implementation details**
→ Read: IMPLEMENTATION_GUIDE.md (entire)

**Make a decision for the project**
→ Read: LOADER_AUTODISCOVERY_SUMMARY.md
→ Then: APPROACHES_COMPARISON.md (first 50 lines)

**Implement this feature**
→ Read: IMPLEMENTATION_GUIDE.md (entire)
→ Reference: GO_PLUGIN_EXPLORATION.md section 9 (naming convention)

---

## Key Findings Summary

### The Problem
HTGO currently requires manual registration of loader functions:
```go
Loaders: map[string]func(c *gin.Context) (any, error){
    "/":      pages.LoadIndex,      // Manual
    "/about": pages.LoadAbout,      // Manual
}
```

New loaders are silently ignored if not manually added - not ideal at scale.

### Options Explored

1. **Go Plugins** - ❌ Not viable
   - Breaks single-binary deployment
   - Platform limited (Linux/macOS only)
   - ABI instability risks

2. **Reflection** - ❌ Problematic
   - Requires architectural changes
   - Loses compile-time safety
   - Runtime overhead

3. **Build-Time Code Generation** - ✓ RECOMMENDED
   - Zero runtime cost
   - Type-safe (compiler checked)
   - Single-binary deployment maintained
   - Aligns with HTGO philosophy

4. **Manual Registration** - ✓ Current approach
   - Works fine but creates boilerplate
   - Doesn't scale well
   - Code generation is clearly better

### The Recommendation

**Build-Time Code Generation with Naming Convention**

Implement `go:generate` integration that:
- Scans `pages/` directory for loader functions
- Derives routes from file paths
- Generates `pages/loaders_generated.go`
- Application uses generated registry instead of manual map

**Naming Convention**:
```
pages/index.go            → func LoadIndex()      → /
pages/about.go            → func LoadAbout()      → /about
pages/blog/[slug].go      → func LoadBlogSlug()   → /blog/:slug
pages/admin/users.go      → func LoadAdminUsers() → /admin/users
```

**Generated Output**:
```go
var LoaderRegistry = map[string]func(c *gin.Context) (any, error){
    "/":           LoadIndex,
    "/about":      LoadAbout,
    "/blog/:slug": LoadBlogSlug,
    "/admin/users": LoadAdminUsers,
}
```

**Application Usage**:
```go
Loaders: pages.LoaderRegistry  // Instead of manual map!
```

---

## Why This Recommendation

1. **Matches HTGO Philosophy**
   - Boring, simple, explicit
   - No runtime magic
   - Predictable behavior

2. **Zero Runtime Cost**
   - All work happens at build time
   - No reflection or dynamic lookups
   - Identical performance to current approach

3. **Type Safe**
   - Generated code checked by compiler
   - Errors caught at build time, not runtime
   - IDE support works perfectly

4. **Platform Neutral**
   - Works on all platforms (unlike plugins)
   - Single-binary deployment maintained
   - Perfect for HTGO's model

5. **Reuses Existing Code**
   - AST parsing already exists in router.go
   - Can extract and reuse validation functions
   - Minimal new code needed

6. **Reversible**
   - Easy to inspect generated code
   - Can override or disable as needed
   - Can roll back if issues arise

---

## Implementation Phases

### Phase 1: Generator Tool (2-3 hours)
- Extract AST utilities from router.go
- Create `cmd/htgo-gen-loaders` standalone tool
- Generates `pages/loaders_generated.go`

### Phase 2: CLI Integration (1-2 hours)
- Add `htgo generate` command
- Auto-run in `htgo dev` and `htgo build`
- Update examples to use generated registry

### Phase 3: Documentation (1 hour)
- Document naming convention in CLAUDE.md
- Add examples and troubleshooting
- Show manual override patterns

---

## Key Statistics

- **Total Documentation**: 1,902 lines
- **Code Examples**: 50+
- **Decision Matrix**: Complete for all approaches
- **Implementation Time Estimate**: 4-6 hours
- **Maintenance Burden**: Minimal (generated code is simple)

---

## Existing HTGO Code Ready for Reuse

The codebase already has most needed infrastructure:

**In `/home/berti/Code/3lines/htgo/router.go`:**
- `DiscoverPages()` (lines 15-69) - finds .tsx files
- `filePathToRoute()` (lines 71-94) - derives routes from paths
- `ListLoaderFiles()` (lines 96-166) - finds colocated .go files
- `isValidLoaderSignature()` (lines 170-201) - validates function types
- `isGinContextType()` (lines 203-220) - type checking utilities
- `isAnyType()` (lines 222-228)
- `isErrorType()` (lines 230-236)

**Usage Pattern:**
```go
// All validation logic can be extracted into reusable module
// Generator tool reuses these functions
// No need to rewrite validation logic
```

---

## Next Steps to Implement

1. **Prototype**: Extract AST utilities to `pkg/loader-analysis/`
2. **Build**: Create `cmd/htgo-gen-loaders/` with complete generator
3. **Test**: Generate registries for both examples
4. **Integrate**: Wire into `htgo dev/build` commands
5. **Document**: Update CLAUDE.md with naming convention
6. **Deploy**: Use in examples as proof of concept

See `IMPLEMENTATION_GUIDE.md` for detailed step-by-step instructions and complete source code.

---

## Questions Answered

**Q: Will this break existing projects?**
A: No. Keep manual registration as an option. Can support both approaches during transition.

**Q: What if I have non-standard function names?**
A: Support manual overrides. Generated registry can be extended in app code.

**Q: Is this extra build step a problem?**
A: No. Auto-integrated into `htgo dev/build` commands.

**Q: Can I verify the generated code?**
A: Yes. Generated file is plain Go, readable and reviewable.

**Q: What about IDE support?**
A: Perfect. Generated functions show in IDE autocomplete, jump-to-definition works.

**Q: Is this reversible?**
A: Yes. Can easily switch back to manual registration or disable generation.

---

## Contact & References

**HTGO Repository**: `/home/berti/Code/3lines/htgo`

**Key Files Referenced**:
- `/home/berti/Code/3lines/htgo/router.go` - AST analysis
- `/home/berti/Code/3lines/htgo/examples/sink/app.go` - Current manual pattern
- `/home/berti/Code/3lines/htgo/types.go` - Type definitions

**Related Go Documentation**:
- `go/ast` - AST parsing
- `go/parser` - Code parsing
- `go:generate` - Build-time code generation

---

## Document Authorship

These reports are comprehensive analysis of Go's capabilities and HTGO's architecture, produced through:
1. Static code analysis of HTGO codebase
2. Research of Go plugin system constraints
3. Evaluation of reflection capabilities
4. Assessment of code generation approaches
5. Trade-off analysis against HTGO's philosophy
6. Implementation planning and examples

All recommendations are grounded in:
- HTGO's stated design philosophy (simplicity, single-binary deployment)
- Go's ecosystem best practices
- Real-world deployment constraints
- Maintainability principles

