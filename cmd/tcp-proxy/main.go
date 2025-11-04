package main

import (
	"log/slog"
	"os"
	"runtime"
	"tcp-proxy/internal/app"
	"tcp-proxy/internal/cli"
)

func main() {
	runtime.GOMAXPROCS(1)

	args := cli.Parse()

	application := app.New(args.Port)
	if err := application.Run(); err != nil {
		slog.Error("Failed to start proxy", slog.Any("err", err))
		os.Exit(1)
	}
}
