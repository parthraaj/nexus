package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	if err := initSVCClient(); err != nil {
		log.Fatalf("failed to initialize SVC client: %v", err)
	}
	log.Printf("SVC client initialized, connected to %s", os.Getenv("SVC_HOST"))

	s := server.NewMCPServer(
		"nexus-storage-mcp-server",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	registerTools(s)

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}

	mux := http.NewServeMux()
	mcpHandler := server.NewStreamableHTTPServer(s)
	mux.Handle("/mcp", mcpHandler)
	mux.Handle("/mcp/", mcpHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("starting nexus-storage-mcp-server v0.1.0 on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
