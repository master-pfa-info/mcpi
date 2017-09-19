// Copyright 2017 The master-pfa-info Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mcpi provides a simple interface to plot (x,y) points
// from a Monte-Carlo approximation method to compute Pi.
//
// Example:
//
//  mcpi.Wait() // wait for the web server to be ready
//  for i := 0; i < 100; i++ {
//      mcpi.Plot(float64(i), float64(i))
//  }
//  mcpi.Wait()
//
package mcpi

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"go-hep.org/x/hep/hplot"
	"golang.org/x/net/websocket"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// Plot plots a point at (x,y)
func Plot(x, y float64) {
	srv.datac <- [2]float64{x, y}
}

// Wait waits for the plot to be finished
func Wait() {
	<-srv.wait
	srv.start = time.Now()
}

// Quit closes the web plot server.
func Quit() {
	log.Printf("total runtime: %v", time.Since(srv.start))
	srv.done <- 1
	<-srv.quit
}

func init() {
	srv = newServer()
}

var (
	srv *server
)

type server struct {
	in  plotter.XYs
	out plotter.XYs
	n   int

	datac chan [2]float64
	plots chan wplot
	quit  chan int
	wait  chan int
	done  chan int
	start time.Time
}

func newServer() *server {
	srv := &server{
		in:    make(plotter.XYs, 0, 1024),
		out:   make(plotter.XYs, 0, 1024),
		datac: make(chan [2]float64),
		plots: make(chan wplot),
		quit:  make(chan int),
		wait:  make(chan int),
		done:  make(chan int),
	}

	go srv.serve()
	go srv.run()

	return srv
}

func (srv *server) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case v := <-srv.datac:
			srv.n++
			x := v[0]
			y := v[1]
			d2 := x*x + y*y
			pt := struct{ X, Y float64 }{x, y}
			switch {
			case d2 < 1:
				srv.in = append(srv.in, pt)
			default:
				srv.out = append(srv.out, pt)
			}
			switch {
			case srv.n < 1e1:
				if srv.n%1e0 == 0 {
					srv.plots <- plot(srv.n, srv.in, srv.out)
				}
			case srv.n < 1e2:
				if srv.n%1e1 == 0 {
					srv.plots <- plot(srv.n, srv.in, srv.out)
				}
			case srv.n < 1e3:
				if srv.n%1e2 == 0 {
					srv.plots <- plot(srv.n, srv.in, srv.out)
				}
			case srv.n < 1e4:
				if srv.n%1e3 == 0 {
					srv.plots <- plot(srv.n, srv.in, srv.out)
				}
			case srv.n < 1e5:
				if srv.n%1e4 == 0 {
					srv.plots <- plot(srv.n, srv.in, srv.out)
				}
			case srv.n < 1e6:
				if srv.n%1e5 == 0 {
					srv.plots <- plot(srv.n, srv.in, srv.out)
				}
			case srv.n < 1e7:
				if srv.n%1e6 == 0 {
					srv.plots <- plot(srv.n, srv.in, srv.out)
				}
			case srv.n > 1e7:
				if srv.n%1e7 == 0 {
					srv.plots <- plot(srv.n, srv.in, srv.out)
				}
			}
		case <-srv.done:
			log.Printf("final: n=%d", srv.n)
			srv.plots <- plot(srv.n, srv.in, srv.out)
			time.Sleep(1 * time.Second) // give the server some time to update
			srv.quit <- 1
			return
		}
	}
}

func plot(n int, in, out plotter.XYs) wplot {
	const pmax = 1e6

	p, err := hplot.New()
	if err != nil {
		log.Fatal(err)
	}

	radius := vg.Points(0.1)

	p.X.Label.Text = "x"
	p.X.Min = 0
	p.X.Max = 1
	p.Y.Label.Text = "y"
	p.Y.Min = 0
	p.Y.Max = 1

	pi := 4 * float64(len(in)) / float64(n)
	p.Title.Text = fmt.Sprintf("n = %d\nπ = %v", n, pi)

	sin, err := hplot.NewScatter(in[:min(pmax, len(in))])
	if err != nil {
		log.Fatal(err)
	}
	sin.Color = color.RGBA{255, 0, 0, 255}
	sin.Radius = radius

	sout, err := hplot.NewScatter(out[:min(pmax/2, len(out))])
	if err != nil {
		log.Fatal(err)
	}
	sout.Color = color.RGBA{0, 0, 255, 255}
	sout.Radius = radius

	p.Add(sin, sout, hplot.NewGrid())

	return wplot{Plot: renderImg(p)}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func renderImg(p *hplot.Plot) string {
	size := 20 * vg.Centimeter
	canvas := vgimg.PngCanvas{vgimg.New(size, size)}
	p.Draw(draw.New(canvas))
	out := new(bytes.Buffer)
	_, err := canvas.WriteTo(out)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(out.Bytes())
}

type wplot struct {
	Plot string `json:"plot"`
}

func (srv *server) serve() {
	port, err := getTCPPort()
	if err != nil {
		log.Fatal(err)
	}
	ip := getIP()
	log.Printf("listening on %s:%s", ip, port)

	http.HandleFunc("/", plotHandle)
	http.Handle("/data", websocket.Handler(dataHandler))
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("error running web-server: %v", err)
	}
}

func plotHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, page)
	select {
	case srv.wait <- 1:
	default:
	}
}

func dataHandler(ws *websocket.Conn) {
	for data := range srv.plots {
		err := websocket.JSON.Send(ws, data)
		if err != nil {
			log.Printf("error sending data: %v\n", err)
		}
	}
}

const page = `
<html>
	<head>
		<title>Monte Carlo</title>
		<script type="text/javascript">
		var sock = null;
		var plot = "";

		function update() {
			var p = document.getElementById("plot");
			p.src = "data:image/png;base64,"+plot;
		};

		window.onload = function() {
			sock = new WebSocket("ws://"+location.host+"/data");

			sock.onmessage = function(event) {
				var data = JSON.parse(event.data);
				plot = data.plot;
				update();
			};
		};

		</script>
	</head>

	<body>
		<div id="content">
			<p style="text-align:center;">
				<img id="plot" src="" alt="Not Available"></img>
			</p>
		</div>
	</body>
</html>
`

func getTCPPort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port), nil
}

func getIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
