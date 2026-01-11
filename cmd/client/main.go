package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	trees "trees"
)

var (
	serverAddr     = flag.String("server", "localhost:9000", "TCP server address")
	projectKey     = flag.String("project", "", "Project key to subscribe to (required)")
	maxConcurrency = flag.Int("concurrency", 5, "Maximum number of concurrent tasks")
	runnerType     = flag.String("runner", "stub", "Runner type: stub or logging")
	sleepDuration  = flag.Duration("sleep", 100*time.Millisecond, "Sleep duration for stub runner")
)

func main() {
	flag.Parse()

	if *projectKey == "" {
		log.Fatal("Error: -project flag is required")
	}

	log.Printf("Starting Trees Client")
	log.Printf("  Server: %s", *serverAddr)
	log.Printf("  Project: %s", *projectKey)
	log.Printf("  Max Concurrency: %d", *maxConcurrency)
	log.Printf("  Runner Type: %s", *runnerType)

	// Create runner based on type
	var runner trees.AgentRunner
	switch *runnerType {
	case "stub":
		log.Printf("  Sleep Duration: %v", *sleepDuration)
		runner = trees.NewStubRunner(*sleepDuration)
	case "logging":
		runner = trees.NewLoggingRunner()
	default:
		log.Fatalf("Unknown runner type: %s (use 'stub' or 'logging')", *runnerType)
	}

	// Create dispatcher
	dispatcher := trees.NewDispatcher(runner, *maxConcurrency)

	// Create client
	client := trees.NewClient(*serverAddr, *projectKey, dispatcher)

	// Setup callback to display summaries
	client.OnTreeReceived = func(summary trees.ExecutionSummary) {
		log.Printf("\n=== EXECUTION SUMMARY ===")
		log.Printf("  Total Tasks: %d", summary.TotalTasks)
		log.Printf("  Successes: %d", summary.Successes)
		log.Printf("  Failures: %d", summary.Failures)
		log.Printf("  Duration: %v", summary.Duration)
		log.Printf("=========================\n")
	}

	// Create context for shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Printf("Received shutdown signal")
		cancel()
	}()

	// Connect and run
	log.Printf("Connecting to server...")
	if err := client.Connect(ctx); err != nil {
		if err != context.Canceled {
			log.Fatalf("Client error: %v", err)
		}
	}

	log.Printf("Client stopped")
}
