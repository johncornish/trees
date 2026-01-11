package main

import (
	"log"
	"net/http"
	"trees/internal/server"
	"trees/internal/store"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	// Start HTTP health server in background
	go func() {
		http.HandleFunc("/health", healthHandler)
		log.Println("HTTP health server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Start Trees TCP server
	store := store.NewStore()
	tcpServer := server.NewTCPServer(":9090", store)
	log.Fatal(tcpServer.Listen())
}
