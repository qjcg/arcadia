# Elbereth Language Specification

**Elbereth** is a practical Lisp dialect that compiles to Go. It combines Lisp's elegance with Go's performance, static types, and concurrency primitives.

---

## Table of Contents

1. [Syntax](#syntax)
2. [Data Types](#data-types)
3. [Type System](#type-system)
4. [Functions & Definitions](#functions--definitions)
5. [Control Flow](#control-flow)
6. [Concurrency](#concurrency)
7. [Go Interoperability](#go-interoperability)
8. [Error Handling](#error-handling)
9. [Macros](#macros)
10. [Standard Library](#standard-library)
11. [Compilation](#compilation)

---

## Syntax

### S-expressions
All code is expressed as S-expressions (symbolic expressions):

```lisp
(function arg1 arg2 arg3)
```

### Special Forms
- `quote` / `'` - Quote an expression (prevent evaluation)
- `quasiquote` / `` ` `` - Quote with unquoting ability
- `unquote` / `,` - Unquote within quasiquote
- `unquote-splicing` / `,@` - Splice in a list

### Comments
```lisp
; line comment

#| block comment |#
```

### Literals

**Numbers:**
```lisp
42              ; integer (int64)
3.14            ; float (float64)
0xFF            ; hex literal
0b1010          ; binary literal
```

**Strings:**
```lisp
"hello"                    ; string (escapes: \n \t \" \\)
```

**Keywords:**
```lisp
:name :age :status         ; keyword (interned symbols)
```

**Booleans & Nil:**
```lisp
true false nil
```

**Collections:**
```lisp
[1 2 3]           ; vector/array
{:name "Alice"}   ; map/object
#{1 2 3}          ; set
```

---

## Data Types

All values have Go runtime representations. Elbereth infers types at compile time.

| Elbereth  | Go Type           | Literal Example                 |
|-----------|-------------------|---------------------------------|
| `int`     | `int64`           | `42`                            |
| `float`   | `float64`         | `3.14`                          |
| `string`  | `string`          | `"hello"`                       |
| `bool`    | `bool`            | `true`                          |
| `rune`    | `rune`            | `\a`                            |
| `keyword` | `string` (tagged) | `:name`                         |
| `[]T`     | `[]T`             | `[1 2 3]`                       |
| `[N]T`    | `[N]T`            | `[3]int{...}`                   |
| `{K V}`   | `map[K]V`         | `{:x 1}`                        |
| `struct`  | Go struct         | `(deftype Point {x int y int})` |
| `chan T`  | `chan T`          | `(chan int)`                    |
| `nil`     | `nil`             | `nil`                           |

### Type Inference Rules
- Numeric literals default to `int` unless context requires `float`
- String literals are always `string`
- Collections infer element types from contents
- Function parameters use explicit or inferred types
- Variable assignments infer from right-hand side

---

## Type System

### Explicit Type Annotations

```lisp
(def x :int 42)
(def y :float 3.14)
(def name :string "Alice")
(def nums :[int] [1 2 3])
```

### Struct Definitions

```lisp
(deftype Point
  {x int
   y int})

(deftype Person
  {name string
   age int
   email :string "default@example.com"})

(def p (Point {x 10 y 20}))
(. p x)  ; => 10
```

### Generic Types

```lisp
(deftype Box :T
  {value T})

(def int-box (Box :int {value 42}))
(def str-box (Box :string {value "hello"}))
```

### Type Aliases

```lisp
(deftype UserId :int)
(def user-id :UserId 123)
```

### Union Types (Sum Type)

```lisp
(deftype Result :T
  (:ok T)
  (:err string))

(def success (Result :int (:ok 42)))
(def failure (Result :int (:err "oops")))
```

---

## Functions & Definitions

### Function Definition

```lisp
(defn add [x y]
  (+ x y))

(defn greet [name :string]
  (str "Hello, " name))

(defn typed-fn [x :int y :int] :int
  (+ x y))
```

### Variadic Functions

```lisp
(defn sum [& args]
  (reduce + 0 args))

(sum 1 2 3 4)  ; => 10
```

### Anonymous Functions (Lambda)

```lisp
(fn [x] (* x 2))

(map (fn [x] (* x 2)) [1 2 3])  ; => [2 4 6]
```

### Function Composition

```lisp
(defn double [x] (* x 2))
(defn inc [x] (+ x 1))

(def pipeline (comp inc double))
(pipeline 5)  ; => 11
```

### First-Class Functions

```lisp
(def add +)
(add 2 3)  ; => 5

(def apply-twice [f x]
  (f (f x)))

(apply-twice double 3)  ; => 12
```

---

## Control Flow

### Conditionals

```lisp
(if condition
  true-branch
  false-branch)

(if (> x 10)
  (println "big")
  (println "small"))
```

### Multiple Branches

```lisp
(cond
  (> x 10) (println "big")
  (> x 5)  (println "medium")
  :else    (println "small"))
```

### Pattern Matching

```lisp
(match status
  :loading    (render-spinner)
  :error msg  (render-error msg)
  :ok data    (render data))

(match value
  (:some x)   (process x)
  :none       (println "nothing"))
```

### Loops

```lisp
; for loop
(for [i 0] (< i 10) [(+ i 1)]
  (println i))

; while loop
(while (< count 10)
  (println count)
  (set! count (+ count 1)))

; reduce/fold
(reduce + 0 [1 2 3 4])  ; => 10

; map/filter/etc
(map (fn [x] (* x 2)) [1 2 3])      ; => [2 4 6]
(filter (fn [x] (> x 2)) [1 2 3 4]) ; => [3 4]
```

### Piping (Threading)

```lisp
(-> value
  (func1)
  (func2 arg)
  (func3))

; Threading with error propagation
(-> result
  (map transform)
  (filter predicate)
  (reduce +))
```

### Conditionals on Options/Results

```lisp
(when (some? value)
  (process value))

(unless (err? result)
  (use-result result))
```

---

## Concurrency

### Goroutines

```lisp
(go (do-work))           ; spawn goroutine, don't wait
(go-wait (do-work))      ; spawn, get channel, wait with <-

(go
  (loop []
    (println "repeating")
    (sleep 1)))
```

### Channels

```lisp
(def ch (chan int))
(def ch (chan string 10))              ; buffered channel

(>! ch 42)                             ; send (blocks if full)
(def val (<! ch))                      ; receive (blocks if empty)

; non-blocking
(>? ch 42)                             ; send, returns false if full
(def [val ok] (<? ch))                 ; receive, ok=false if empty
```

### Channel Select

```lisp
(select
  [ch1 val1] (handle-ch1 val1)
  [ch2 val2] (handle-ch2 val2)
  [default]  (handle-timeout))

; select with timeout
(select
  [ch1 val] (println val)
  [(after 1000) _] (println "timeout"))
```

### Wait Groups

```lisp
(let [wg (sync/WaitGroup)]
  (wg.Add 3)
  (go (work) (wg.Done))
  (go (work) (wg.Done))
  (go (work) (wg.Done))
  (wg.Wait))
```

### Mutex & Synchronization

```lisp
(deftype Counter
  {mu sync/Mutex
   val int})

(defn inc [c]
  (.Lock c.mu)
  (defer (.Unlock c.mu))
  (set! c.val (+ c.val 1)))
```

---

## Go Interoperability

### Importing Go Packages

```lisp
(import "encoding/json")
(import "fmt")
(import "net/http")

; aliased import
(import [json "encoding/json"])
```

### Calling Go Functions

```lisp
(json/Marshal my-struct)           ; package function
(fmt/Println "hello")

; calling methods
(.Write writer "data")             ; (receiver.method args)
(.Len my-slice)
```

### Go Structs & Types

```lisp
(deftype Request
  {method string
   path string
   body :string ""})

; construct
(def req (Request {method "GET" path "/api"}))

; access fields
(. req method)

; modify fields
(set! (. req body) "new body")
```

### Converting Go Interfaces

```lisp
; io.Writer interface automatically satisfied by methods
(deftype MyWriter []
  (Write [b :[byte]] :int))

(defn my-write [b :[byte]] :int
  (len b))

; use as io.Writer
(io/WriteString my-writer "hello")
```

### Embedding External Go

For complex Go interop, embed Go directly:

```lisp
(go-code
  "func customLogic() string {
     return fmt.Sprintf(\"custom: %d\", 42)
   }")

(def result (customLogic))
```

---

## Error Handling

### Result Type

```lisp
(deftype Result :T
  (:ok T)
  (:err string))

(defn risky-op [] (Result :int)
  (:err "something failed"))

(defn maybe-process [r (Result :int)]
  (match r
    (:ok val)  (println "success:" val)
    (:err msg) (println "error:" msg)))
```

### Try-Catch

```lisp
(try
  (risky-operation)
  :catch err
  (handle-error err))
```

### Defer

```lisp
(defn read-file [path]
  (let [file (open-file path)]
    (defer (close file))
    (read file)))
```

### Error Propagation

```lisp
(-> result
  (result-map do-thing1)
  (result-map do-thing2)
  (result-or-else handle-error))
```

---

## Macros

### Macro Definition

```lisp
(defmacro unless [condition & body]
  `(if (not ~condition)
     (do ~@body)))

(unless (ready?)
  (println "not ready yet"))
```

### Quote/Unquote

```lisp
(defmacro log-val [x]
  `(do
     (println "value:" ~x)
     ~x))

(log-val (+ 1 2))  ; => "value: 3" then 3
```

### Syntax Quotes

```lisp
(defmacro incr [x]
  `(+ ~x 1))
```

---

## Standard Library

### Arithmetic

```lisp
(+ 1 2 3)      ; => 6
(- 10 3)       ; => 7
(* 2 3)        ; => 6
(/ 10 3)       ; => 3.33
(mod 10 3)     ; => 1
(pow 2 8)      ; => 256
(sqrt 16)      ; => 4.0
(abs -5)       ; => 5
```

### Comparison

```lisp
(= 1 1)        ; => true
(!= 1 2)       ; => true
(< 1 2)        ; => true
(<= 1 1)       ; => true
(> 2 1)        ; => true
(>= 2 2)       ; => true
```

### Logic

```lisp
(and true false)     ; => false
(or false true)      ; => true
(not true)           ; => false
```

### String Operations

```lisp
(str "Hello" " " "World")      ; => "Hello World"
(len "hello")                   ; => 5
(upper "hello")                 ; => "HELLO"
(lower "HELLO")                 ; => "hello"
(trim "  hello  ")              ; => "hello"
(split "a,b,c" ",")             ; => ["a" "b" "c"]
(join ["a" "b" "c"] ",")        ; => "a,b,c"
(starts-with "hello" "hel")     ; => true
(contains "hello" "ll")         ; => true
```

### Collections

```lisp
(len [1 2 3])              ; => 3
(first [1 2 3])            ; => 1
(rest [1 2 3])             ; => [2 3]
(last [1 2 3])             ; => 3
(nth [1 2 3] 1)            ; => 2
(conj [1 2] 3)             ; => [1 2 3]
(cons 0 [1 2])             ; => [0 1 2]
(concat [1 2] [3 4])       ; => [1 2 3 4]
(reverse [1 2 3])          ; => [3 2 1]
(map func coll)            ; apply func to each element
(filter pred coll)         ; keep elements where pred is true
(reduce func init coll)    ; fold from left
(some pred coll)           ; true if any element matches
(every pred coll)          ; true if all elements match
(get coll key)             ; map/slice access
(get coll key default)     ; with default
(keys map)                 ; all keys
(values map)               ; all values
```

### Type Checking

```lisp
(type? :int 42)               ; => true
(type? :string "hello")        ; => true
(some? value)                  ; => true (not nil)
(none? value)                  ; => true (is nil)
(err? result)                  ; => true (is :err type)
(ok? result)                   ; => true (is :ok type)
```

### Conversions

```lisp
(int "42")                 ; => 42
(float "3.14")             ; => 3.14
(string 42)                ; => "42"
(bytes "hello")            ; => [104 101 108 108 111]
(keyword "name")           ; => :name
```

### I/O

```lisp
(println "hello")                    ; print with newline
(print "hello")                      ; print without newline
(println "name:" name "age:" age)    ; multiple args

(read-file "path.txt")               ; => string or error
(write-file "path.txt" "content")    ; writes and returns error

(input)                              ; read line from stdin
(input "prompt: ")                   ; read with prompt
```

### JSON

```lisp
(import [json "encoding/json"])

(json/marshal obj)              ; => json string or error
(json/unmarshal json-str type)  ; => object or error

(def data {
  :name "Alice"
  :age 30
})

(def json-str (json/marshal data))
(def parsed (json/unmarshal json-str Person))
```

### HTTP (builtin helpers)

```lisp
(import [http "net/http"])

(http.Get "https://example.com")
(http.Post "https://example.com" "application/json" body)

(defn handler [w http/ResponseWriter req *http/Request]
  (.Header w).("Content-Type") "application/json"
  (fmt/Fprintf w "{}"))

(http.HandleFunc "/api" handler)
(http.ListenAndServe ":8080" nil)
```

### Testing

```lisp
(deftest add-test
  (assert-eq (add 2 3) 5)
  (assert-true (> 5 3))
  (assert-false (< 5 3)))

(deftest error-test
  (let [result (failing-op)]
    (assert-err result)))
```

### Time & Concurrency Helpers

```lisp
(import [time "time"])

(now)                      ; => current time
(sleep 1000)               ; => sleep 1000ms
(sleep-until (after 5))    ; => wait max 5s

(after 1000)               ; => channel that fires after 1s
(ticker 100)               ; => channel that fires every 100ms
(timeout 5000)             ; => chan that fires after 5s
```

---

## Compilation

### Build Process

```
Source (.elb)
  → Lexer (tokenize)
  → Parser (build AST)
  → Type Resolution (infer & check types)
  → Go Code Generation
  → go build
  → Binary
```

### Command-Line Interface

```bash
elbereth init <module>        # initialize a module (runs go mod init)
elbereth build <path>         # compile package(s) to binary or Go code
elbereth run <path>           # compile and run main package
elbereth test <path>          # compile and run tests
elbereth tidy                 # tidy dependencies (runs go mod tidy)
elbereth repl                 # start interactive REPL
elbereth check <path>         # syntax and type check
```

### Main Entry Point

```lisp
(defn main []
  (println "Hello, Elbereth!"))
```

Compiles to Go's `func main()`.

### Packaging

Elbereth uses Go's package system directly. A file declares its package using the `package` form:

```lisp
(package myapp)

(defn ExportedFunc [] ...)  ; Exported (starts with uppercase)
(defn privateFunc [] ...)   ; Private (starts with lowercase)
```

Module paths and dependencies are managed via `go.mod`.

### Build Configuration

Elbereth directly uses `go.mod`, `go.sum`, and `go.work` files exactly like Go. Dependencies should be managed via `elbereth get` or `elbereth tidy` (which wrap `go get` and `go mod tidy`).

---

## Examples

### Hello World

```lisp
(defn main []
  (println "Hello, World!"))
```

### HTTP Server

```lisp
(import [http "net/http"])
(import [fmt "fmt"])

(defn handler [w http/ResponseWriter req *http/Request]
  (.Header w).("Content-Type") "application/json"
  (fmt/Fprintf w "{\"message\": \"Hello\"}"))

(defn main []
  (http.HandleFunc "/api" handler)
  (println "Server starting on :8080")
  (http.ListenAndServe ":8080" nil))
```

### Concurrent Data Processing

```lisp
(defn process-items [items]
  (let [results (chan [:itemid int :value int] (len items))]
    (doseq [item items]
      (go
        (let [result (* item.value 2)]
          (>! results {:itemid item.id :value result}))))

    (collect-results results (len items))))

(defn collect-results [ch count]
  (loop [acc [] i 0]
    (if (< i count)
      (recur (conj acc (<! ch)) (+ i 1))
      acc)))
```

### Error Handling Pattern

```lisp
(deftype User {id int name string})

(defn find-user [id :int] (Result :User)
  (let [user (db-lookup id)]
    (if (some? user)
      (:ok user)
      (:err "user not found"))))

(defn get-user-handler [w http/ResponseWriter req *http/Request]
  (let [id (parse-id req)
        result (find-user id)]
    (match result
      (:ok user)
      (do
        (fmt/Fprintf w (json/marshal user)))
      (:err msg)
      (do
        (.WriteHeader w http.StatusNotFound)
        (fmt/Fprintf w (json/marshal {:error msg}))))))
```

---

## Notes

- **Compilation Target**: Go 1.21+
- **No Runtime**: Compiles directly to Go binaries; no Elbereth runtime
- **Type Safety**: Static types checked at compile time; no runtime type errors
- **Performance**: Equivalent to hand-written Go (goroutines, channels, etc.)
- **Interop**: Seamlessly call Go libraries; no FFI boundaries
- **Tooling**: REPL for interactive development; formatter for consistent style
