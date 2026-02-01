# Elbereth

A practical Lisp dialect that compiles to Go.

Elbereth combines the elegance and expressiveness of Lisp with Go's performance, concurrency primitives, and simplicity. Write beautiful, functional code that compiles to fast, efficient Go binaries.

## Why Elbereth?

- **Lisp Expressiveness**: Clean S-expressions, macros, first-class functions
- **Go Performance**: Compiles to native Go; no VM overhead
- **Native Concurrency**: Goroutines, channels, and mutexes as first-class language features
- **Static Types**: Full type inference with Go-like safety
- **Seamless Interop**: Call Go libraries directly without FFI boundaries
- **Single Binary**: No runtime dependencies; just ship the binary

## Quick Example

```lisp
(import [http "net/http"])

(defn handler [w http/ResponseWriter req *http/Request]
  (fmt/Fprintf w "Hello, %s!" (.URL req).Path))

(defn main []
  (http.HandleFunc "/" handler)
  (http.ListenAndServe ":8080" nil))
```

Compiles to idiomatic Go code and runs with Go's performance.

## Documentation

See [SPEC.md](./SPEC.md) for the complete language specification covering:

- **Syntax & Literals**: S-expressions, numbers, strings, collections
- **Type System**: Inference, structs, generics, union types
- **Functions**: Definition, composition, variadic args, closures
- **Control Flow**: Conditionals, pattern matching, loops, threading
- **Concurrency**: Goroutines, channels, select, wait groups
- **Go Interop**: Importing packages, calling functions, structs
- **Error Handling**: Result types, try-catch, defer, error propagation
- **Macros**: Meta-programming with quote/unquote
- **Standard Library**: Math, strings, collections, I/O, JSON, HTTP
- **Compilation**: Build process, CLI tools, module system

## Project Status

🚧 **Specification Phase** - Language design and API exploration

This is the formal specification for the Elbereth language. Implementation will follow.

## Design Philosophy

1. **Practical over Pure**: Real-world applicability trumps theoretical purity
2. **Go-Native**: Compile to Go, not to an intermediate; use Go's ecosystem directly
3. **Explicit over Implicit**: Clear error handling, type annotations when needed
4. **Fast Development**: REPL, hot-reload, clear error messages
5. **Concurrent First**: Make parallelism easy and correct

## Inspiration

- **Clojure**: Practical Lisp, immutability, strong data structures
- **Go**: Simplicity, performance, built-in concurrency, fast compilation
- **Scheme**: Minimalist core, macros, first-class continuations
- **Racket**: Excellent tooling, pattern matching, contracts

## Use Cases

- **Microservices**: Goroutine-backed HTTP handlers with elegant syntax
- **CLI Tools**: Fast, small binaries with clean argument parsing
- **Data Processing**: Functional pipelines with type safety
- **Game Servers**: Concurrent networking with channels
- **Systems Software**: Performance-critical applications

