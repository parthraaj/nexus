package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	if err := initK8sClient(); err != nil {
		log.Fatalf("failed to initialize k8s client: %v", err)
	}

	s := server.NewMCPServer(
		"nexus-node-mcp-server",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	registerTools(s)

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}

	// Mount MCP handler at /mcp and health at /health on same mux
	mux := http.NewServeMux()
	mcpHandler := server.NewStreamableHTTPServer(s)
	mux.Handle("/mcp", mcpHandler)
	mux.Handle("/mcp/", mcpHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("starting nexus-node-mcp-server v0.1.0 on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
