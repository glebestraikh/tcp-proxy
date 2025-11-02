package cli

import (
	"flag"
	"fmt"
	"os"
)

type Args struct {
	Port int
}

func Parse() *Args {
	port := flag.Int("port", 0, "Server port (required)")
	flag.Parse()

	if *port == 0 {
		fmt.Println("Missing required flag: -port")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *port < 1 || *port > 65535 {
		fmt.Printf("Port %d is out of range (1-65535)\n", *port)
		os.Exit(1)
	}

	return &Args{Port: *port}
}
