# testscript Tests for Terebra

## Setup

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

### Already Implemented (`basic.txtar`)
- Basic execution: `echo hello`
- Brace expansion: `{a,b,c}`, `{1..5}`
- Arithmetic: `$((2 + 3))`
- Indexed arrays: `arr=(a b c) ; echo ${arr[1]}`
- Associative arrays: `colors[red]="#ff0000" ; echo ${colors[red]}`
- Key listing: `${!colors[@]}`
- Exit codes: `! exec terebra -c 'exit 1'`

### 1. Variables (`vars.txtar`)

```
# variable expansion
exec terebra -c 'echo $HOME'
stdout '/home/'

# local variable assignment
exec terebra -c 'FOO=bar ; echo $FOO'
stdout 'bar'

# PS1 default
exec terebra -c 'echo $PS1'
stdout 'trb:'

# $$ PID
exec terebra -c 'echo $$'
stdout '[0-9]+' PANE

# $? exit code
exec terebra -c 'false ; echo $?'
stdout '1'

# ${VAR} braces
exec terebra -c 'FOO=bar ; echo ${FOO}'
stdout 'bar'

# undefined variable
exec terebra -c 'echo $UNDEFINED_VAR_12345'
stdout ''
```

### 2. String Manipulation (`strings.txtar`)

```
# substring
exec terebra -c 'FOO=hello ; echo ${FOO:0:2}'
stdout 'he'

# replace first
exec terebra -c 'FOO=hello ; echo ${FOO/l/xyz}'
stdout 'hexyzo'

# replace all
exec terebra -c 'FOO=hello ; echo ${FOO//l/xyz}'
stdout 'hexyzyzo'

# uppercase first
exec terebra -c 'FOO=hello ; echo ${FOO^}'
stdout 'Hello'

# uppercase all
exec terebra -c 'FOO=hello ; echo ${FOO^^}'
stdout 'HELLO'

# length
exec terebra -c 'FOO=hello ; echo ${#FOO}'
stdout '5'
```

### 3. Command Chaining (`chaining.txtar`)

```
# && — success
exec terebra -c 'true && echo ok'
stdout 'ok'

# && — failure (short-circuit)
exec terebra -c 'false && echo nope'
! stdout 'nope'

# || — fallthrough
exec terebra -c 'false || echo ok'
stdout 'ok'

# || — short-circuit on success
exec terebra -c 'true || echo nope'
! stdout 'nope'

# ;
exec terebra -c 'echo a ; echo b'
stdout 'a'
stdout 'b'
```

### 4. Pipes and Redirects (`pipes.txtar`)

```
# pipe
exec terebra -c 'echo hello | wc -c'
stdout '6'

# redirect stdout
exec terebra -c 'echo hello > /tmp/terebra_test_out ; cat /tmp/terebra_test_out'
stdout 'hello'

# redirect append
exec terebra -c 'echo a > /tmp/terebra_test_app ; echo b >> /tmp/terebra_test_app ; cat /tmp/terebra_test_app'
stdout 'a'
stdout 'b'

# redirect stdin
exec terebra -c 'echo hello > /tmp/terebra_test_in ; cat < /tmp/terebra_test_in'
stdout 'hello'
```

### 5. Globbing (`glob.txtar`)

```
-- hello.txt --
hello
-- world.txt --
world

exec terebra echo *.txt
stdout 'hello.txt'
stdout 'world.txt'

# recursive glob
exec terebra echo **/*.txt
stdout 'hello.txt'
stdout 'world.txt'

# no match — passes literal
exec terebra echo *.nonexistent
stdout '*.nonexistent'
```

### 6. Heredocs (`heredoc.txtar`)

```
# basic heredoc
exec terebra -c 'cat << EOF\nhello\nworld\nEOF'
stdout 'hello'
stdout 'world'

# heredoc with variable expansion
exec terebra -c 'NAME=terebra\ncat << EOF\nhello $NAME\nEOF'
stdout 'hello terebra'

# heredoc with quoted delimiter (no expansion)
exec terebra -c 'NAME=terebra\ncat << '"'"'EOF'"'"'\nhello $NAME\nEOF'
stdout 'hello $NAME'
```

### 7. Command Substitution (`cmdsubst.txtar`)

```
# $(cmd)
exec terebra -c 'echo $(echo hello)'
stdout 'hello'

# nested $(cmd)
exec terebra -c 'echo $(echo $(echo deep))'
stdout 'deep'

# backtick
exec terebra -c 'echo `echo hello`'
stdout 'hello'
```

### 8. Arrays (`arrays.txtar`)

```
# indexed array — all values
exec terebra -c 'arr=(a b c) ; echo ${arr[@]}'
stdout 'a b c'

# indexed array — length
exec terebra -c 'arr=(a b c) ; echo ${#arr[@]}'
stdout '3'

# indexed array — element assignment
exec terebra -c 'arr=(a b c) ; arr[1]=x ; echo ${arr[1]}'
stdout 'x'

# associative array — inline init
exec terebra -c 'declare -A colors ; colors=([red]=r [blue]=b) ; echo ${colors[red]}'
stdout 'r'

# associative array — all values
exec terebra -c 'colors[red]=r ; colors[blue]=b ; echo ${colors[@]}'
stdout 'r b'

# array auto-promote (non-numeric key)
exec terebra -c 'arr[key]=val ; echo ${arr[key]}'
stdout 'val'
```

### 9. Builtins (`builtins.txtar`)

```
# type
exec terebra -c 'type echo'
stdout 'echo is a shell builtin'

# type external
exec terebra -c 'type ls'
stdout 'ls is '

# which
exec terebra -c 'which echo'
stdout '/bin/echo'

# export
exec terebra -c 'export FOO=bar ; echo $FOO'
stdout 'bar'

# unset
exec terebra -c 'FOO=bar ; unset FOO ; echo $FOO'
stdout ''

# readonly
exec terebra -c 'readonly FOO=bar ; FOO=baz ; echo $FOO'
stdout 'bar'

# alias
exec terebra -c 'alias ll="ls -la" ; alias'
stdout 'll='

# unalias
exec terebra -c 'alias ll="ls -la" ; unalias ll ; alias'
! stdout 'll'

# set -x debug trace
exec terebra -c 'set -x ; echo hello'
stderr '+ echo hello'

# set +x disable debug
exec terebra -c 'set -x ; set +x ; echo hello'
! stderr 'hello'
```

### 10. Script Files (`script.txtar`)

```
-- test.trb --
echo "hello from script"

exec terebra test.trb
stdout 'hello from script'

# script with shebang
-- shebang.trb --
#!/usr/bin/env terebra
echo "shebang works"

exec terebra shebang.trb
stdout 'shebang works'
```

### 11. CUE Integration (`cue.txtar`)

```
-- data.cue --
name: "test"
value: 42

exec terebra -c 'cue eval data.cue'
stdout 'name:'
stdout 'value: 42'

# cue export as JSON
exec terebra -c 'cue export data.cue'
stdout '"name"'
stdout '"test"'
stdout '42'

# drill cue
exec terebra -c 'drill cue data.cue'
stdout 'name:'
stdout 'value: 42'

# drill cue with extract
exec terebra -c 'drill cue data.cue -e name'
stdout '"test"'
```

### 12. Help and Flags (`help.txtar`)

```
# help
exec terebra help
stdout 'Terebra -- auger shell'

# --version
exec terebra --version
stdout 'terebra 0.1.0'

# --explain
exec terebra --explain echo hello
stdout '# would execute: echo hello'
```

### 13. Shell Options (`options.txtar`)

```
# set -o vi
exec terebra -c 'set -o vi ; echo hello'
stdout 'hello'

# set -o emacs
exec terebra -c 'set -o emacs ; echo hello'
stdout 'hello'
```

### 14. Error Cases (`errors.txtar`)

```
# unknown command
! exec terebra -c 'nonexistent_command_xyz'
stderr 'command not found'

# parse error
! exec terebra -c '|'
stderr 'parse error'

# type not found
! exec terebra -c 'type nonexistent_cmd'
stderr 'not found'
```

### 15. Source (`source.txtar`)

```
-- lib.trb --
echo "sourced"

exec terebra -c 'source lib.trb'
stdout 'sourced'
```

## Execution

```bash
go test ./cmd/terebra/ -run TestTerebra -v
```

## Notes

- Test scripts run in a temp directory — files declared with `-- filename --` are created there
- Each `exec terebra ...` runs a fresh terebra instance (script mode), not the REPL
- `-c` flag passes inline script to terebra
- Use `\$` to prevent shell expansion in the testscript runner
- `!` prefix on `exec` means the command is expected to fail
- `stdout 'pattern'` matches if the pattern appears anywhere in stdout
- `! stdout 'pattern'` asserts the pattern does NOT appear
- `stderr 'pattern'` matches on stderr
- `stdout '[0-9]+' PANE` uses regex matching with PANE flag
