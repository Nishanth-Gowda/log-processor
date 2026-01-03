package processor

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	json "github.com/goccy/go-json"
)

// OffsetData represents the persisted offset state for a segment
type OffsetData struct {
	Segment        string    `json:"segment"`
	Offset         int64     `json:"offset"`
	LinesProcessed int64     `json:"lines_processed"`
	LastUpdated    time.Time `json:"last_updated"`
}

// OffsetManager manages offsets for log segments
type OffsetManager struct {
	offsetDir string
	offsets   map[string]*OffsetData
	mu        sync.RWMutex
}

// NewOffsetManager creates a new offset manager
func NewOffsetManager(offsetDir string) (*OffsetManager, error) {
	if err := os.MkdirAll(offsetDir, 0755); err != nil {
		return nil, err
	}

	om := &OffsetManager{
		offsetDir: offsetDir,
		offsets:   make(map[string]*OffsetData),
	}

	// Load existing offsets
	if err := om.loadAll(); err != nil {
		return nil, err
	}

	return om, nil
}

// loadAll loads all offset files from disk
func (om *OffsetManager) loadAll() error {
	files, err := filepath.Glob(filepath.Join(om.offsetDir, "*.offset.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip unreadable files
		}

		var offset OffsetData
		if err := json.Unmarshal(data, &offset); err != nil {
			continue // Skip corrupted files
		}

		om.offsets[offset.Segment] = &offset
	}

	return nil
}

// GetOffset returns the last committed offset for a segment
func (om *OffsetManager) GetOffset(segment string) (int64, int64) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	if data, ok := om.offsets[segment]; ok {
		return data.Offset, data.LinesProcessed
	}
	return 0, 0
}

// CommitOffset saves the offset for a segment
func (om *OffsetManager) CommitOffset(segment string, offset int64, linesProcessed int64) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	data := &OffsetData{
		Segment:        segment,
		Offset:         offset,
		LinesProcessed: linesProcessed,
		LastUpdated:    time.Now().UTC(),
	}

	om.offsets[segment] = data

	// Persist to disk
	return om.persist(segment, data)
}

// persist writes offset data to disk
func (om *OffsetManager) persist(segment string, data *OffsetData) error {
	filename := filepath.Join(om.offsetDir, segment+".offset.json")

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first, then rename (atomic)
	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, filename)
}

// IsComplete checks if a segment has been fully processed
func (om *OffsetManager) IsComplete(segment string, fileSize int64) bool {
	om.mu.RLock()
	defer om.mu.RUnlock()

	if data, ok := om.offsets[segment]; ok {
		return data.Offset >= fileSize
	}
	return false
}

// GetAllOffsets returns all tracked offsets
func (om *OffsetManager) GetAllOffsets() map[string]OffsetData {
	om.mu.RLock()
	defer om.mu.RUnlock()

	result := make(map[string]OffsetData)
	for k, v := range om.offsets {
		result[k] = *v
	}
	return result
}
