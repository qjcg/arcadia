# Benthos websocket example

You can explore the websocket connection with various tools including:

- [websocat](https://github.com/vi/websocat)

## Example (websocat)

```sh
# Run an interactive websocket session
$ websocat ws://localhost:4195/ws

# Test a single string
$ echo hello benthos | websocat ws://localhost:4195/ws
```
