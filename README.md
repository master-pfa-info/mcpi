# mcpi

[![GoDoc](https://godoc.org/github.com/master-pfa-info/mcpi?status.svg)](https://godoc.org/github.com/master-pfa-info/mcpi)

`mcpi` provides a simple interface to draw `(x,y)` points for a Monte-Carlo approximation method to compute Pi.

`mcpi` starts a web server that will listen for plot requests.

## Installation

Installation is done with `go get`:

```sh
$> go get github.com/master-pfa-info/mcpi
```

## Example

```go
func main() {
	mcpi.Wait()
	mcpi.Plot(0,0)
	mcpi.Plot(0.5, 0.5)
	mcpi.Quit()
}
```

```sh
$> go run ./main.go
2017/09/19 17:27:44 listening on 127.0.0.1:46191
```

and then direct your favorite web-browser to the indicated URL.

## Sample

![mc-pi](https://github.com/master-pfa-info/mcpi/raw/master/mc-pi.png)
