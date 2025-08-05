The barr command prints out a status line for use with minimalistic window managers.

# Features

- simple
- lightweight
- easy to extend (blocks simply need to implement the [fmt.Stringer](https://golang.org/pkg/fmt/#Stringer) interface)


# Usage

```shell
# print a status line to stdout
barr
```

## With i3 or Sway

Add a `status_command` line like this to your [i3](https://i3wm.org/) or [Sway](https://swaywm.org) config file (adjusting the sleep value as needed):

```
bar {
	...
	status_command while barr; do sleep 5 ; done
}
```


# External Tool Dependencies

The `barr` command itself does not depend on any external tools, but some individual blocks do.

Specifically, the blocks below have external tool dependencies:

- volume: depends on `pactl`
- wifi: depends on `iw`
