# mcpi

[![GoDoc](https://godoc.org/github.com/master-pfa-info/mcpi?status.svg)](https://godoc.org/github.com/master-pfa-info/mcpi)

`mcpi` provides a simple interface to draw `(x,y)` points for a Monte-Carlo approximation method to compute Pi.

## Installation

Installation is done with `go get`:

```sh
$> go get github.com/master-pfa-info/mcpi
```

## Example

```go
mcpi.Wait()
mcpi.Plot(0,0)
mcpi.Plot(0.5, 0.5)
mcpi.Quit()
```
