package processor

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// SegmentState represents the processing state of a segment
type SegmentState int

const (
	SegmentPending    SegmentState = iota // Ready for processing
	SegmentProcessing                     // Being processed by a worker
	SegmentComplete                       // Fully processed
)

// Segment represents a log file segment
type Segment struct {
	Name     string       // Segment filename (e.g., "app.log.20260101-231106")
	Path     string       // Full path to segment file
	Size     int64        // File size in bytes
	State    SegmentState // Current processing state
	WorkerID int          // Assigned worker ID (-1 if unassigned)
}

// SegmentManager manages log file segments
type SegmentManager struct {
	logsDir   string
	pattern   string // Base log file pattern (e.g., "app.log")
	segments  map[string]*Segment
	offsetMgr *OffsetManager
	mu        sync.RWMutex
}

// NewSegmentManager creates a new segment manager
func NewSegmentManager(logsDir, pattern string, offsetMgr *OffsetManager) *SegmentManager {
	return &SegmentManager{
		logsDir:   logsDir,
		pattern:   pattern,
		segments:  make(map[string]*Segment),
		offsetMgr: offsetMgr,
	}
}

// Scan discovers all available segments in the logs directory
func (sm *SegmentManager) Scan() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Find all rotated log files (pattern.TIMESTAMP format)
	files, err := filepath.Glob(filepath.Join(sm.logsDir, sm.pattern+".*"))
	if err != nil {
		return err
	}

	for _, path := range files {
		// Skip offset files
		if strings.HasSuffix(path, ".offset.json") || strings.HasSuffix(path, ".tmp") {
			continue
		}

		name := filepath.Base(path)

		// Skip if already tracked
		if _, exists := sm.segments[name]; exists {
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		// Determine state based on offset
		state := SegmentPending
		if sm.offsetMgr.IsComplete(name, info.Size()) {
			state = SegmentComplete
		}

		sm.segments[name] = &Segment{
			Name:     name,
			Path:     path,
			Size:     info.Size(),
			State:    state,
			WorkerID: -1,
		}
	}

	return nil
}

// GetPendingSegments returns segments ready for processing
func (sm *SegmentManager) GetPendingSegments() []*Segment {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var pending []*Segment
	for _, seg := range sm.segments {
		if seg.State == SegmentPending {
			pending = append(pending, seg)
		}
	}

	// Sort by name (chronological order)
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Name < pending[j].Name
	})

	return pending
}

// ClaimSegment attempts to claim a segment for a worker
func (sm *SegmentManager) ClaimSegment(segmentName string, workerID int) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	seg, exists := sm.segments[segmentName]
	if !exists || seg.State != SegmentPending {
		return false
	}

	seg.State = SegmentProcessing
	seg.WorkerID = workerID
	return true
}

// MarkComplete marks a segment as fully processed
func (sm *SegmentManager) MarkComplete(segmentName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if seg, exists := sm.segments[segmentName]; exists {
		seg.State = SegmentComplete
		seg.WorkerID = -1
	}
}

// ReleaseSegment releases a segment back to pending (e.g., on worker failure)
func (sm *SegmentManager) ReleaseSegment(segmentName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if seg, exists := sm.segments[segmentName]; exists {
		seg.State = SegmentPending
		seg.WorkerID = -1
	}
}

// GetSegment returns a segment by name
func (sm *SegmentManager) GetSegment(name string) *Segment {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.segments[name]
}

// GetStats returns segment statistics
func (sm *SegmentManager) GetStats() (total, pending, processing, complete int) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, seg := range sm.segments {
		total++
		switch seg.State {
		case SegmentPending:
			pending++
		case SegmentProcessing:
			processing++
		case SegmentComplete:
			complete++
		}
	}
	return
}
