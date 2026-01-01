package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log-processor/internal/logger"
)

func main() {
	// Command line flags
	interval := flag.Duration("interval", 500*time.Millisecond, "Interval between log generation")
	format := flag.String("format", "json", "Output format: json or text")
	count := flag.Int("count", 0, "Number of logs to generate (0 for infinite)")
	flag.Parse()

	fmt.Println("🚀 Log Generator Started")
	fmt.Printf("   Interval: %v\n", *interval)
	fmt.Printf("   Format: %s\n", *format)
	if *count > 0 {
		fmt.Printf("   Count: %d\n", *count)
	} else {
		fmt.Println("   Count: infinite (Ctrl+C to stop)")
	}
	fmt.Println("---")

	// Create logging service
	svc := logger.NewService("log-generator")

	// Setup graceful shutdown
	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n✋ Shutting down...")
		close(done)
	}()

	// Generate logs
	logChan := make(chan logger.LogEntry, 100)
	go svc.GenerateLogs(*interval, logChan, done)

	generated := 0
	for {
		select {
		case entry := <-logChan:
			if *format == "json" {
				fmt.Println(entry.FormatJSON())
			} else {
				fmt.Println(entry.FormatText())
			}

			generated++
			if *count > 0 && generated >= *count {
				log.Printf("Generated %d logs. Exiting.", generated)
				return
			}
		case <-done:
			log.Printf("Generated %d logs total.", generated)
			return
		}
	}
}
