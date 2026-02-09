# Bun Examples

This directory contains simple examples of using Bun for various tasks.

## Examples

1.  **CLI Tool** (`cli/`): A simple command-line interface that echoes text and can transform it to uppercase.
2.  **Web Server** (`server/`): A basic HTTP server using `Bun.serve`.
3.  **Fullstack URL Shortener** (`fullstack/`): A complete example with a frontend (HTML/JS), backend (API), and SQLite database (`bun:sqlite`).

## Working with Examples

First, install dependencies:

```bash
bun install
```

### Running Examples

-   **CLI**: `bun cli/index.ts "hello world" --upper`
-   **Server**: `bun server/index.ts` (Listen on http://localhost:3000)
-   **Fullstack**: `bun fullstack/index.ts` (Listen on http://localhost:3001)

### Running Tests

All examples include tests using Bun's built-in test runner:

```bash
bun test
```
