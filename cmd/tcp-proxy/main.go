package main

import (
	"runtime"

	"tcp-proxy/internal/cli"
	"tcp-proxy/internal/proxy"
)

func main() {
	runtime.GOMAXPROCS(1)

	args := cli.Parse()
	proxy.Start(args.Port)
}
