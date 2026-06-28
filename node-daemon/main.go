package main

import (
	"log"
	"os"
)

func main() {
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "0.0.0.0:9090"
	}
	log.Printf("starting nexus-node-daemon v0.1.0")
	startServer(addr)
}
