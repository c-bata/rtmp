package main

import (
	"flag"
	"log"
	"os"

	"github.com/c-bata/rtmp"
)

var (
	revision string
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":1935", `proxy local address`)
	flag.Parse()

	log.Printf("Serving RTMP on %s (rev-%s)", addr, revision)
	err := rtmp.ListenAndServe(addr)
	if err != nil {
		log.Printf("Got Error: %s", err)
		os.Exit(1)
	}
}
