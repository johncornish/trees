package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	trees "trees"
)

var (
	tcpAddr  = flag.String("tcp", ":9000", "TCP address for client connections")
	httpAddr = flag.String("http", ":8080", "HTTP address for API and health check")
)

func main() {
	flag.Parse()

	log.Printf("Starting Trees Server")
	log.Printf("  TCP address: %s (for client connections)", *tcpAddr)
	log.Printf("  HTTP address: %s (for API and health check)", *httpAddr)

	// Create TCP server
	server := trees.NewServer(*tcpAddr)

	// Start TCP server in background
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("TCP server error: %v", err)
		}
	}()

	// Setup HTTP handlers
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/publish", publishHandler(server))

	// Start HTTP server in background
	go func() {
		log.Printf("HTTP server listening on %s", *httpAddr)
		if err := http.ListenAndServe(*httpAddr, nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Printf("Shutting down...")
	server.Stop()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func publishHandler(server *trees.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var tree trees.Tree
		if err := json.NewDecoder(r.Body).Decode(&tree); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Publish the tree to subscribed clients
		server.PublishTree(tree)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "published",
			"treeId": tree.ID,
		})
	}
}
