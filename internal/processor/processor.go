package processor

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Config holds the processor configuration
type Config struct {
	LogsDir      string
	LogPattern   string
	OffsetsDir   string
	WorkerCount  int
	ScanInterval time.Duration
}

// ProcessFunc is the callback function for processing each log record
type ProcessFunc func(*LogRecord) error

// Processor orchestrates log file processing
type Processor struct {
	cfg         Config
	processFunc ProcessFunc

	offsetMgr  *OffsetManager
	segmentMgr *SegmentManager

	workers  []*worker
	workerWg sync.WaitGroup

	processed atomic.Int64
	errors    atomic.Int64

	ctx     context.Context
	cancel  context.CancelFunc
	running atomic.Bool
}

// worker processes segments
type worker struct {
	id        int
	processor *Processor
}

// NewProcessor creates a new log processor
func NewProcessor(cfg Config, processFunc ProcessFunc) (*Processor, error) {
	// Create offset manager
	offsetMgr, err := NewOffsetManager(cfg.OffsetsDir)
	if err != nil {
		return nil, err
	}

	// Create segment manager
	segmentMgr := NewSegmentManager(cfg.LogsDir, cfg.LogPattern, offsetMgr)

	p := &Processor{
		cfg:         cfg,
		processFunc: processFunc,
		offsetMgr:   offsetMgr,
		segmentMgr:  segmentMgr,
	}

	// Create workers
	p.workers = make([]*worker, cfg.WorkerCount)
	for i := 0; i < cfg.WorkerCount; i++ {
		p.workers[i] = &worker{
			id:        i,
			processor: p,
		}
	}

	return p, nil
}

// Start begins processing log files
func (p *Processor) Start(ctx context.Context) error {
	if p.running.Swap(true) {
		return nil // Already running
	}

	p.ctx, p.cancel = context.WithCancel(ctx)

	// Initial scan
	if err := p.segmentMgr.Scan(); err != nil {
		return err
	}

	// Start workers
	for _, w := range p.workers {
		p.workerWg.Add(1)
		go w.run()
	}

	// Start scanner goroutine
	go p.scanLoop()

	return nil
}

// Stop gracefully stops the processor
func (p *Processor) Stop() {
	if !p.running.Swap(false) {
		return // Not running
	}

	if p.cancel != nil {
		p.cancel()
	}

	// Wait for workers to finish
	p.workerWg.Wait()
}

// Stats returns processing statistics
func (p *Processor) Stats() (processed, errors int64, segmentStats [4]int) {
	processed = p.processed.Load()
	errors = p.errors.Load()

	total, pending, processing, complete := p.segmentMgr.GetStats()
	segmentStats = [4]int{total, pending, processing, complete}

	return
}

// scanLoop periodically scans for new segments
func (p *Processor) scanLoop() {
	ticker := time.NewTicker(p.cfg.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			_ = p.segmentMgr.Scan()
		}
	}
}

// run is the main loop for a worker
func (w *worker) run() {
	defer w.processor.workerWg.Done()

	for {
		select {
		case <-w.processor.ctx.Done():
			return
		default:
			// Get pending segments
			segments := w.processor.segmentMgr.GetPendingSegments()
			if len(segments) == 0 {
				// No work, sleep briefly
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Try to claim a segment
			for _, seg := range segments {
				if w.processor.segmentMgr.ClaimSegment(seg.Name, w.id) {
					w.processSegment(seg)
					break
				}
			}
		}
	}
}

// processSegment processes a single segment
func (w *worker) processSegment(seg *Segment) {
	// Get starting offset
	startOffset, _ := w.processor.offsetMgr.GetOffset(seg.Name)

	// Create reader
	reader, err := NewLogReader(seg.Path, startOffset)
	if err != nil {
		w.processor.errors.Add(1)
		w.processor.segmentMgr.ReleaseSegment(seg.Name)
		return
	}
	defer reader.Close()

	var linesProcessed int64

	// Process each record
	for {
		select {
		case <-w.processor.ctx.Done():
			// Save progress before exiting
			_ = w.processor.offsetMgr.CommitOffset(seg.Name, reader.Offset(), linesProcessed)
			w.processor.segmentMgr.ReleaseSegment(seg.Name)
			return
		default:
		}

		record, err := reader.Read()
		if err != nil {
			// EOF or error - mark complete
			break
		}

		// Process the record
		if err := w.processor.processFunc(record); err != nil {
			w.processor.errors.Add(1)
		} else {
			w.processor.processed.Add(1)
			linesProcessed++
		}

		// Commit offset periodically (every 100 records)
		if linesProcessed%100 == 0 {
			_ = w.processor.offsetMgr.CommitOffset(seg.Name, reader.Offset(), linesProcessed)
		}
	}

	// Final offset commit
	_ = w.processor.offsetMgr.CommitOffset(seg.Name, reader.Offset(), linesProcessed)
	w.processor.segmentMgr.MarkComplete(seg.Name)
}
