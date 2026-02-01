# Elbereth Getting Started Guide

Welcome to Elbereth – a practical Lisp that compiles to Go!

## Installation

```bash
cd x/elbereth
go build -o elbereth ./cmd/elbereth
sudo mv elbereth /usr/local/bin/
```

## Quick Start

### Hello World

Create `hello.elb`:

```lisp
(defn main []
  (println "Hello, Elbereth!"))
```

Run it:

```bash
elbereth run hello.elb
```

Output:
```
Hello, Elbereth!
```

## Basic Concepts

### Functions

Define functions with `defn`:

```lisp
(defn add [x y]
  (+ x y))

(defn main []
  (println (add 2 3)))  ; => 5
```

### Variables

Define top-level variables with `def`:

```lisp
(def pi 3.14159)

(def message "Hello")

(defn main []
  (println message))
```

### Conditionals

Use `if` for branching:

```lisp
(defn is-positive [n]
  (if (> n 0)
    "positive"
    "not positive"))

(defn main []
  (println (is-positive 5)))   ; => positive
```

### Operators

Arithmetic and comparison:

```lisp
(+ 1 2 3)          ; => 6
(- 10 3)           ; => 7
(* 2 3 4)          ; => 24
(/ 10 2)           ; => 5
(> 5 3)            ; => true
(< 5 3)            ; => false
(== 5 5)           ; => true
```

### Collections

Vectors (arrays) and maps:

```lisp
(def numbers [1 2 3])
(def person {:name "Alice" :age 30})

(defn main []
  (println numbers)   ; => [1 2 3]
  (println person))   ; => {:name Alice :age 30}
```

### Printing

Use `println` and `print`:

```lisp
(defn main []
  (println "Hello")      ; prints with newline
  (print "World")        ; prints without newline
  (println "!"))         ; => HelloWorld!
```

## Examples

### Example 1: Function Definition

```lisp
(defn factorial [n]
  (if (== n 0)
    1
    (* n (factorial (- n 1)))))

(defn main []
  (println "5! =" (factorial 5)))
```

### Example 2: Multiple Arguments

```lisp
(defn greet [greeting person]
  (println greeting ", " person "!"))

(defn main []
  (greet "Hello" "World")
  (greet "Hi" "Elbereth"))
```

### Example 3: Top-Level Definitions

```lisp
(def name "Elbereth")
(def version "0.1.0")

(defn show-version []
  (println name "v" version))

(defn main []
  (show-version))
```

## Compilation

### Generating Go Code

```bash
elbereth build program.elb -o program
```

This generates a compiled binary.

### Type Checking

```bash
elbereth check program.elb
```

This parses and checks your program without compiling.

## Standard Library

### String Functions

```lisp
(println "Hello")           ; print with newline
(print "text")              ; print without newline
```

### Arithmetic

```lisp
(+ x y)             ; addition
(- x y)             ; subtraction
(* x y)             ; multiplication
(/ x y)             ; division
(% x y)             ; modulo
```

### Comparison

```lisp
(== x y)            ; equal
(!= x y)            ; not equal
(< x y)             ; less than
(<= x y)            ; less than or equal
(> x y)             ; greater than
(>= x y)            ; greater than or equal
```

### Logic

```lisp
(and a b)           ; logical AND
(or a b)            ; logical OR
(not a)             ; logical NOT
```

## Advanced Features (Coming Soon)

### Structs

```lisp
(deftype Person
  {name string
   age int})
```

### Concurrency

```lisp
(def ch (chan int))
(go (println "Goroutine!"))
(>! ch 42)
(def val (<! ch))
```

### Error Handling

```lisp
(try
  (risky-operation)
  :catch err
  (handle-error err))
```

### Macros

```lisp
(defmacro unless [cond & body]
  `(if (not ~cond)
     (do ~@body)))
```

## Project Structure

```
program.elb               ; your Elbereth source file
program                   ; compiled binary (after build)
```

## Troubleshooting

### "Unknown token"
Check your syntax. Elbereth uses Lisp S-expressions.

### Compilation error
Run `elbereth check program.elb` to see type and syntax errors.

### Program won't run
Make sure you have a `main` function defined:
```lisp
(defn main []
  (println "Entry point"))
```

## Tips & Tricks

### Use Multiple Functions

Break your code into small functions:

```lisp
(defn helper [x]
  (+ x 1))

(defn process [data]
  (helper data))

(defn main []
  (println (process 5)))
```

### Comments

Use `;` for line comments:

```lisp
; This is a comment
(println "Hello")  ; end-of-line comment

; Block comments
#| Multi-line
   block comment |#
```

### Printing Multiple Values

```lisp
(println "x =" x "y =" y)   ; multiple arguments
```

## Next Steps

1. Read the full [SPEC.md](./SPEC.md) for language details
2. Explore [examples/](./examples/) for more code samples
3. Try building small programs and incrementally add features
4. Join the community (coming soon!)

## Version

Elbereth v0.1.0 (Alpha)

Status: Function definitions, basic arithmetic, printing, conditionals working.
Future: Structs, concurrency, macros, full stdlib.

