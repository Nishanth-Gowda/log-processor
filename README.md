<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"/>
</p>

<h1 align="center">⚡ Log Processor</h1>

<p align="center">
  <strong>A high-performance, resumable log processing system written in Go</strong>
</p>

<p align="center">
  <em>Built for speed. Designed for reliability. Made for scale.</em>
</p>

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🚀 **High Performance** | Uses [`goccy/go-json`](https://github.com/goccy/go-json) for blazing-fast JSON parsing |
| 📁 **Segment-Based Processing** | Handles rotated log files with automatic discovery |
| 💾 **Resumable Processing** | Persists byte offsets to disk — never reprocess data |
| 👷 **Worker Pool** | Configurable parallel workers for concurrent processing |
| 🔄 **Log Rotation Support** | Seamlessly handles rotating log files (1MB segments) |
| 🛑 **Graceful Shutdown** | Saves progress on SIGINT/SIGTERM for safe restarts |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Log Processor                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐      │
│   │   Scanner   │────▶│  Dispatcher │────▶│   Workers   │      │
│   │   (1s tick) │     │             │     │   (N pool)  │      │
│   └─────────────┘     └─────────────┘     └──────┬──────┘      │
│                                                   │             │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │                  Segment Manager                         │  │
│   │  ┌─────────┐  ┌────────────┐  ┌──────────┐              │  │
│   │  │ Pending │─▶│ Processing │─▶│ Complete │              │  │
│   │  └─────────┘  └────────────┘  └──────────┘              │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                                                 │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │                  Offset Manager                          │  │
│   │  • Persists byte offsets to disk                        │  │
│   │  • Enables resumable processing                         │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📦 Project Structure

```
log-processor/
├── cmd/
│   ├── generator/          # Log generation tool
│   │   └── main.go
│   └── processor/          # Main log processor
│       └── main.go
├── internal/
│   ├── logger/             # Log entry structures & generation
│   │   ├── logger.go
│   │   └── logger_bench_test.go
│   └── processor/          # Core processing engine
│       ├── processor.go    # Main orchestrator
│       ├── segment.go      # Segment discovery & management
│       ├── reader.go       # Log file reader with offset tracking
│       └── offset.go       # Offset persistence
├── logs/                   # Generated log files (gitignored)
├── offsets/                # Offset tracking files (gitignored)
├── scripts/
│   └── clean.sh
├── Makefile
└── README.md
```

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/log-processor.git
cd log-processor

# Install dependencies
make deps
```

### Generate Test Logs

```bash
# Generate 10,000 log entries
make generate-test-logs

# Or run the generator directly
go run cmd/generator/main.go -count 10000 -interval 1ms
```

### Run the Processor

```bash
# Using make
make run-processor

# Or run directly with options
go run cmd/processor/main.go -workers 4 -logs-dir logs -pattern app.log
```

---

## ⚙️ Configuration

### Processor Options

| Flag | Default | Description |
|------|---------|-------------|
| `-logs-dir` | `logs` | Directory containing log files |
| `-pattern` | `app.log` | Base log file pattern |
| `-offsets-dir` | `offsets` | Directory for offset files |
| `-workers` | `2` | Number of parallel workers |

### Generator Options

| Flag | Default | Description |
|------|---------|-------------|
| `-count` | `1000` | Number of log entries to generate |
| `-interval` | `10ms` | Interval between log entries |
| `-output` | `logs` | Output directory |

---

## 📊 Log Format

The processor handles JSON log entries with the following structure:

```json
{
  "timestamp": "2026-01-02T12:30:45.123456789Z",
  "level": "INFO",
  "service": "api-gateway",
  "message": "Request completed",
  "request_id": "req-a1b2c3d4",
  "user_id": "user-1234",
  "duration_ms": 245
}
```

### Log Levels

| Level | Weight | Description |
|-------|--------|-------------|
| `DEBUG` | 15% | Detailed debugging information |
| `INFO` | 50% | General operational messages |
| `WARNING` | 20% | Warning conditions |
| `ERROR` | 10% | Error conditions |
| `FATAL` | 5% | Critical failures |

---

## 🛠️ Make Commands

```bash
make build           # Build all binaries
make generator       # Build generator only
make processor       # Build processor only
make run-generator   # Run the log generator
make run-processor   # Run the log processor
make test            # Run all tests
make bench           # Run benchmarks
make clean           # Clean build artifacts
make clean-logs      # Delete generated logs
make clean-offsets   # Delete offset files
make clean-all       # Clean everything
make reset           # Full reset for fresh testing
```

---

## 📈 Performance

Benchmarked on Apple Silicon (M-series):

| Operation | Throughput | Notes |
|-----------|------------|-------|
| JSON Parsing | ~500K ops/sec | Using goccy/go-json |
| Log Processing | ~50K records/sec | Single worker |
| Log Processing | ~150K records/sec | 4 workers |

---

## 🔄 Resumable Processing

The processor automatically saves progress:

1. **Offset files** are stored in `offsets/` as JSON:
   ```json
   {
     "segment": "app.log.20260102-122240",
     "offset": 524377,
     "lines_processed": 3000,
     "last_updated": "2026-01-02T08:45:18Z"
   }
   ```

2. **On restart**, processing resumes from the last committed offset
3. **Offsets commit** every 100 records for durability

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

<p align="center">
  Made with ❤️ and Go
</p>
