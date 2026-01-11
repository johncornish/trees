package trees

import (
	"context"
	"log"
	"sync"
	"time"
)

// Dispatcher orchestrates the parallel execution of tasks
type Dispatcher struct {
	runner         AgentRunner
	maxConcurrency int
}

// NewDispatcher creates a new task dispatcher
func NewDispatcher(runner AgentRunner, maxConcurrency int) *Dispatcher {
	return &Dispatcher{
		runner:         runner,
		maxConcurrency: maxConcurrency,
	}
}

// Dispatch executes all tasks in a tree in parallel, respecting concurrency limits
func (d *Dispatcher) Dispatch(ctx context.Context, tree Tree) ExecutionSummary {
	start := time.Now()

	log.Printf("[DISPATCHER] Starting dispatch for tree %s (project: %s) with %d tasks",
		tree.ID, tree.ProjectKey, len(tree.Tasks))

	if len(tree.Tasks) == 0 {
		log.Printf("[DISPATCHER] No tasks to execute")
		return ExecutionSummary{
			TotalTasks: 0,
			Successes:  0,
			Failures:   0,
			Duration:   time.Since(start),
			Results:    []TaskResult{},
		}
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, d.maxConcurrency)

	// Channel to collect results
	results := make(chan TaskResult, len(tree.Tasks))

	// WaitGroup to wait for all tasks to complete
	var wg sync.WaitGroup

	// Launch all tasks
	for _, task := range tree.Tasks {
		wg.Add(1)

		go func(t TaskNode) {
			defer wg.Done()

			// Acquire semaphore slot
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				// Context cancelled before we could start
				results <- TaskResult{
					TaskID:   t.ID,
					Success:  false,
					Error:    ctx.Err(),
					Duration: 0,
				}
				return
			}

			// Execute the task
			result, err := d.runner.Run(ctx, t)
			if err != nil {
				result.Error = err
				result.Success = false
			}

			results <- result
		}(task)
	}

	// Wait for all tasks to complete
	wg.Wait()
	close(results)

	// Collect and summarize results
	summary := ExecutionSummary{
		TotalTasks: len(tree.Tasks),
		Duration:   time.Since(start),
		Results:    make([]TaskResult, 0, len(tree.Tasks)),
	}

	for result := range results {
		summary.Results = append(summary.Results, result)
		if result.Success {
			summary.Successes++
		} else {
			summary.Failures++
		}
	}

	log.Printf("[DISPATCHER] Completed dispatch in %v: %d successes, %d failures",
		summary.Duration, summary.Successes, summary.Failures)

	return summary
}
