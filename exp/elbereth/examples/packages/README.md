# Elbereth Package Example

This directory illustrates the Elbereth package and module system.

## Structure

- `lib/math.elb`: A library package (`package lib`) containing math functions.
- `lib/lib_test.elb`: Unit tests for the `lib` package.
- `main.elb`: A main package (`package main`) that imports and uses `lib`.

## Running

1. **Build the library**:
   ```bash
   elbereth build examples/packages/lib
   ```
   This generates `examples/packages/lib/math.go`.

2. **Run the main program**:
   ```bash
   elbereth run examples/packages/main.elb
   ```

3. **Run tests**:
   ```bash
   elbereth test examples/packages/lib
   ```
