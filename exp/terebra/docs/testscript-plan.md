# testscript Tests for Terebra

## Setup

Add testscript dependency and create test harness:

```go
// cmd/terebra/main_test.go
package main

import (
    "testing"
    "github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
    testscript.Main(m, map[string]func(){
        "terebra": main,
    })
}

func TestTerebra(t *testing.T) {
    testscript.Run(t, testscript.Params{
        Dir: "testdata",
    })
}
```

## Test Scripts

### 1. Basic Execution

```go
// testdata/basic.txtar
# simple command
exec terebra echo hello
stdout 'hello'

# exit code
! exec terebra exit 1
```

### 2. Variables

```go
// testdata/vars.txtar
# variable expansion
exec terebra echo \$HOME
stdout '/home/'

# local variable assignment
exec terebra sh -c 'echo $FOO'
! stdout '.'

# PS1 default
exec terebra sh -c 'echo $PS1'
stdout 'trb:'
```

### 3. Command Chaining

```go
// testdata/chaining.txtar
# &&
exec terebra sh -c 'true && echo ok'
stdout 'ok'

# ||
! exec terebra sh -c 'false && echo nope'
! stdout 'nope'

# ;
exec terebra sh -c 'echo a ; echo b'
stdout 'a'
stdout 'b'
```

### 4. Pipes and Redirects

```go
// testdata/pipes.txtar
# pipe
exec terebra sh -c 'echo hello | wc -c'
stdout '6'

# redirect stdout
exec terebra sh -c 'echo hello > /tmp/terebra_test_out && cat /tmp/terebra_test_out'
stdout 'hello'
```

### 5. Brace Expansion

```go
// testdata/brace.txtar
exec terebra echo {a,b,c}
stdout 'a b c'

exec terebra echo {1..5}
stdout '1 2 3 4 5'
```

### 6. Globbing

```go
// testdata/glob.txtar
-- hello.txt --
hello
-- world.txt --
world

exec terebra ls *.txt
stdout 'hello.txt'
stdout 'world.txt'
```

### 7. Heredocs

```go
// testdata/heredoc.txtar
exec terebra sh -c 'cat << EOF\nhello\nworld\nEOF'
stdout 'hello'
stdout 'world'
```

### 8. Arrays

```go
// testdata/arrays.txtar
# indexed array
exec terebra sh -c 'arr=(a b c) ; echo ${arr[1]}'
stdout 'b'

# associative array
exec terebra sh -c 'arr[key]=val ; echo ${arr[key]}'
stdout 'val'
```

### 9. Command Substitution

```go
// testdata/cmdsubst.txtar
# $(cmd)
exec terebra echo \$(echo hello)
stdout 'hello'

# backtick
exec terebra echo \`echo hello\`
stdout 'hello'
```

### 10. Arithmetic

```go
// testdata/arithmetic.txtar
exec terebra echo \$((2 + 3))
stdout '5'
```

### 11. Aliases

```go
// testdata/alias.txtar
exec terebra sh -c 'alias ll="ls -la" && alias'
stdout 'll='
```

### 12. Script Execution

```go
// testdata/script.txtar
-- test.trb --
echo "hello from script"
exec terebra test.trb
stdout 'hello from script'
```

### 13. CUE Integration

```go
// testdata/cue.txtar
-- data.cue --
name: "test"
value: 42
exec terebra cue eval data.cue
stdout 'name:'
stdout 'value: 42'
```

## Execution

```bash
go test ./cmd/terebra/ -run TestTerebra -v
```

## Notes

- Test scripts run in a temp directory — files declared with `-- filename --` are created there
- Each `exec terebra ...` runs a fresh terebra instance (script mode), not the REPL
- `sh -c '...'` is needed for compound commands since terebra script mode runs one line at a time
- Use `\$` to prevent shell expansion in the testscript runner
- `!` prefix on `exec` means the command is expected to fail
