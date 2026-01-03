package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log-processor/internal/processor"
)

func main() {
	// Command line flags
	logsDir := flag.String("logs-dir", "logs", "Directory containing log files")
	pattern := flag.String("pattern", "app.log", "Base log file pattern")
	offsetsDir := flag.String("offsets-dir", "offsets", "Directory for offset files")
	workers := flag.Int("workers", 2, "Number of parallel workers")
	flag.Parse()

	fmt.Println("Log Processor Started")
	fmt.Printf("Logs Dir: %s\n", *logsDir)
	fmt.Printf("Pattern: %s\n", *pattern)
	fmt.Printf("Offsets Dir: %s\n", *offsetsDir)
	fmt.Printf("Workers: %d\n", *workers)
	fmt.Println("---")

	// Create processor configuration
	cfg := processor.Config{
		LogsDir:      *logsDir,
		LogPattern:   *pattern,
		OffsetsDir:   *offsetsDir,
		WorkerCount:  *workers,
		ScanInterval: time.Second,
	}

	// Example process function - just count by level
	levelCounts := make(map[string]int64)
	var totalCount int64

	processFunc := func(record *processor.LogRecord) error {
		totalCount++
		level := string(record.Entry.Level)
		if level != "" {
			levelCounts[level]++
		}

		// Print every 1000th record as progress
		if totalCount%1000 == 0 {
			fmt.Printf("Processed: %d records", totalCount)
		}

		return nil
	}

	// Create processor
	proc, err := processor.NewProcessor(cfg, processFunc)
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Start processing
	if err := proc.Start(ctx); err != nil {
		log.Fatalf("Failed to start processor: %v", err)
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Stop processor
	proc.Stop()

	// Print final stats
	processed, errors, segStats := proc.Stats()
	fmt.Println("\n\nFinal Statistics")
	fmt.Printf("Total Processed: %d\n", processed)
	fmt.Printf("Errors: %d\n", errors)
	fmt.Printf("Segments - Total: %d, Pending: %d, Processing: %d, Complete: %d\n",
		segStats[0], segStats[1], segStats[2], segStats[3])

	fmt.Println("\nLog Levels:")
	for level, count := range levelCounts {
		fmt.Printf("   %s: %d\n", level, count)
	}
}
