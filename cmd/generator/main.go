package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"log-processor/internal/logger"
)

func main() {
	// Command line flags
	interval := flag.Duration("interval", 500*time.Millisecond, "Interval between log generation")
	format := flag.String("format", "json", "Output format: json or text")
	count := flag.Int("count", 0, "Number of logs to generate (0 for infinite)")
	output := flag.String("output", "logs/app.log", "Output log file path")
	rotate := flag.Int64("rotate-size", 10, "Rotate log file when it reaches this size in MB (0 to disable)")
	flag.Parse()

	// Create output directory if it doesn't exist
	dir := filepath.Dir(*output)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Open log file
	file, err := os.OpenFile(*output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	fmt.Println("🚀 Log Generator Started")
	fmt.Printf("   Output: %s\n", *output)
	fmt.Printf("   Interval: %v\n", *interval)
	fmt.Printf("   Format: %s\n", *format)
	if *rotate > 0 {
		fmt.Printf("   Rotate at: %d MB\n", *rotate)
	}
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
	rotateBytes := *rotate * 1024 * 1024 // Convert MB to bytes
	var currentSize int64

	// Get current file size
	if info, err := file.Stat(); err == nil {
		currentSize = info.Size()
	}

	for {
		select {
		case entry := <-logChan:
			var line string
			if *format == "json" {
				line = entry.FormatJSON()
			} else {
				line = entry.FormatText()
			}

			// Write to file
			n, err := writer.WriteString(line + "\n")
			if err != nil {
				log.Printf("Error writing to log file: %v", err)
				continue
			}
			currentSize += int64(n)

			generated++

			// Flush periodically for visibility
			if generated%100 == 0 {
				writer.Flush()
				fmt.Printf("\r📝 Generated %d logs (%.2f MB)", generated, float64(currentSize)/(1024*1024))
			}

			// Check for rotation
			if rotateBytes > 0 && currentSize >= rotateBytes {
				writer.Flush()
				file.Close()

				// Rotate file
				timestamp := time.Now().Format("20060102-150405")
				rotatedName := fmt.Sprintf("%s.%s", *output, timestamp)
				os.Rename(*output, rotatedName)
				fmt.Printf("\n🔄 Rotated log to: %s\n", rotatedName)

				// Open new file
				file, err = os.OpenFile(*output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if err != nil {
					log.Fatalf("Failed to open new log file: %v", err)
				}
				writer = bufio.NewWriter(file)
				currentSize = 0
			}

			if *count > 0 && generated >= *count {
				writer.Flush()
				fmt.Printf("\n✅ Generated %d logs to %s\n", generated, *output)
				return
			}

		case <-done:
			writer.Flush()
			fmt.Printf("\n✅ Generated %d logs total to %s\n", generated, *output)
			return
		}
	}
}
