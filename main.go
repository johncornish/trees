package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"trees/api"
	"trees/graph"
)

func main() {
	dataDir := os.Getenv("TREES_DATA_DIR")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		dataDir = filepath.Join(home, ".trees")
	}
	storePath := filepath.Join(dataDir, "data.json")

	handler, err := api.NewHandler(storePath, &graph.ExecGitChecker{})
	if err != nil {
		log.Fatal(err)
	}

	addr := os.Getenv("TREES_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("Server starting on %s (data: %s)", addr, storePath)
	if err := http.ListenAndServe(addr, handler.Mux()); err != nil {
		log.Fatal(err)
	}
}
