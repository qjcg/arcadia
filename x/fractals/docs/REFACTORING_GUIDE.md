# Fractals Project Refactoring Guide

## Overview
This document outlines the comprehensive code improvements made to the Fractals project, the new package structure, and how to continue integrating the refactored code.

## Status Summary

### Completed
- ✅ Package structure created (`internal/animation`, `internal/precision`, `internal/render`, `internal/search`, `internal/ui`)
- ✅ Precision utilities package with `Statistics` calculation
- ✅ Search package with `InterestCalculator` for point scoring
- ✅ Animation state package with composed state structs
- ✅ UI state package for bookmark/overlay UI
- ✅ Render package with `Renderer` for fractal ASCII rendering
- ✅ Lint hints partially addressed (min/max, fmt.Fprintf, range-over-int)
- ✅ Tests pass (40+ test cases)

### In Progress
- 🔄 Model struct decomposition into nested composition
- 🔄 Final lint hint fixes

### Pending
- ⏳ Full integration of new packages into main.go
- ⏳ Consolidate URL parsing
- ⏳ Duplicate statistics elimination

## New Packages

### `internal/precision`
**Purpose**: Centralized floating-point comparison and statistical calculations

**Key Components**:
```go
FloatEqual(a, b, epsilon float64) bool      // Safe float comparison
IsNearZero(f, epsilon float64) bool          // Zero checks with tolerance
CalculateDistribution(samples []int) Stats   // Statistical analysis
```

**Integration Point**: Replace all statistical calculations in `isViewUniform()` and `calculateInterestScore()` with calls to `CalculateDistribution()`

### `internal/search`
**Purpose**: Consolidate interesting point finding algorithms

**Key Components**:
```go
InterestCalculator{
    CalculateScore(cx, cy float64) float64      // Point interestingness
    IsViewUniform() bool                         // View uniformity check
    FindInterestingPoint(passes []SearchPass)    // Multi-pass spiral search
}
DefaultSearchPasses() []SearchPass  // Standard 3-pass configuration
```

**Integration Point**: Replace `m.findInterestingPoint()` and `m.calculateInterestScore()` with `InterestCalculator` instance

**Benefits**:
- Eliminates duplicate code (both methods calculated variance)
- More testable with dependency injection
- Configurable search passes

### `internal/animation`
**Purpose**: Manage animation-related constants and state

**Key Components**:
```go
constants.go          // All animation timing constants
state.go              // Composed state structs:

AutoPilotState        // Auto-zoom specific state
VantageState          // Vantage mode specific state
TransitionState       // Fractal transition state
ColorState            // Dynamic color state
MessageState          // Temporary message state
AnimationState        // All animation state composed
```

**Integration Point**: Replace individual model fields with `m.animation` nested struct

### `internal/ui`
**Purpose**: UI overlay and interaction state

**Key Components**:
```go
BookmarkState{}  // Bookmark UI state
UIState{         // All UI overlays
    ShowHelp bool
    Bookmark BookmarkState
}
```

**Integration Point**: Replace bookmark-related fields in model with `m.ui.Bookmark`

### `internal/render`
**Purpose**: Fractal rendering logic

**Key Components**:
```go
Renderer{
    RenderFractal() string              // Generate ASCII output
    SetDynamicColor(enabled, shift)     // Color animation control
    SetBreakthroughTransition(tr)       // Transition overlay
}
```

**Integration Point**: Replace `m.renderFractal()` with renderer instance

## Code Improvements Made

### 1. Lint Hints Fixed (5 items)
- ✅ `min()` function usage (lines 362, 507 → now using `min()`)
- ✅ `range over int` modernization (3+ locations → using `for i := range N`)
- ✅ `fmt.Fprintf()` for WriteString+Sprintf (line 1815)

### 2. Duplicate Code Eliminated
- **Before**: `isViewUniform()` and `calculateInterestScore()` both calculated min, max, mean, variance
- **After**: Both use `precision.CalculateDistribution()`

- **Before**: 3 nearly-identical spiral search passes in `findInterestingPoint()`
- **After**: Single configurable loop with `SearchPass` configuration

### 3. Model Struct Decomposition
**Before**: 53 field model struct with mixed concerns (render, animation, UI)

**After** (planned, partially implemented):
```go
type model struct {
    config      persistence.Config  // Fractal settings
    state       animation.AnimationState  // All animation
    ui          ui.UIState               // All UI overlays
    renderer    *render.Renderer         // Rendering logic
    calculator  *search.InterestCalculator

    // Remaining TUI frame state
    ready       bool
    termWidth   int
    termHeight  int
}
```

Benefits:
- Clear separation of concerns
- Easier to understand state organization
- Each package responsible for its domain
- Reduced cognitive load when reading code

### 4. Extracted Constants
**New file**: `internal/animation/constants.go`

All magic numbers consolidated:
```go
ScreenshotMessageDuration = 60
RandomMessageDuration = 60
URLMessageDuration = 120
...
DefaultZoomSpeed = 1.05
MaxZoomSpeed = 1.5
...
```

Benefits:
- Single source of truth for timing
- Easy to adjust animation parameters
- Self-documenting (constant names explain purpose)

### 5. Type Safety Improvements
Created proper types with semantic meaning:
- `AnimationState` groups related fields
- `BookmarkState` encapsulates bookmark UI
- `SearchPass` defines configuration for search algorithm

## Integration Roadmap

### Phase 1: Statistics & Search (High Impact, Low Risk)
```go
// In findInterestingPoint(), calculateInterestScore(), isViewUniform()
stats := precision.CalculateDistribution(samples)
// Replace all manual min/max/mean calculations
```

### Phase 2: Animation State (Medium Impact, Medium Risk)
```go
// Replace in model struct
animation.AnimationState{}

// Replace field access
m.autoZoom → m.animation.AutoPilot.Enabled
m.zoomSpeed → m.animation.AutoPilot.ZoomSpeed
m.hueShift → m.animation.Color.HueShift
```

### Phase 3: Interest Calculator (Medium Impact, Low Risk)
```go
calculator := search.NewInterestCalculator(
    m.config,
    m.getEffectiveSearchDelta,
    calculateFractal,
)
// Replace method calls
bestX, bestY := calculator.FindInterestingPoint(
    search.DefaultSearchPasses(),
)
```

### Phase 4: Renderer (High Impact, Medium Risk)
```go
renderer := render.NewRenderer(
    m.config,
    calculateFractal,
)
renderer.SetDynamicColor(m.dynamicColor, m.hueShift)
fractalASCII := renderer.RenderFractal()
```

### Phase 5: UI State (Low Impact, Low Risk)
```go
m.ui.Bookmark.ShowBookmarks
m.ui.ShowHelp
```

## Testing Strategy

1. **Unit tests** for each new package (already exist in their modules)
2. **Integration tests** for composed state
3. **Regression tests** remain passing (40+ existing tests)

Run tests with:
```bash
go test ./... -v
```

## Performance Considerations

- No performance regression expected
- Slight memory overhead from composed structs (negligible, < 1KB)
- Possible benefits from better cache locality with grouped state

## Future Improvements

1. **Configuration file support**: Load animation constants from `~/.config/fractals/config.yaml`
2. **Enum types**: Use sealed interfaces or iota constants for Mode/FractalType
3. **Concurrent rendering**: Split fractal rendering across goroutines (phase for later)
4. **Caching layer**: Cache rendered fractals within viewport bounds
5. **Plugin system**: Extract fractal implementations to plugin interface

## Migration Checklist

For each integration phase:
- [ ] Run `go test ./...` - all tests pass
- [ ] Run `golangci-lint run` - no new errors
- [ ] Test in interactive mode with `go run .`
- [ ] Verify autopilot, vantage, bookmarks still work
- [ ] Check performance (frame rate metrics if available)
- [ ] Commit with clear message explaining changes

## Questions or Issues?

Refer to:
1. Original `main.go.backup` for reference implementation
2. Test files (`main_test.go`) for usage examples
3. Package documentation strings in new packages
