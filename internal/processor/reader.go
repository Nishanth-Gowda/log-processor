package processor

import (
	"bufio"
	"io"
	"os"

	"log-processor/internal/logger"

	json "github.com/goccy/go-json"
)

// LogReader reads log entries from a segment with offset tracking
type LogReader struct {
	file       *os.File
	reader     *bufio.Reader
	segment    string
	offset     int64 // Current byte offset
	lineNumber int64 // Current line number
}

// NewLogReader creates a reader for a segment, starting from the given offset
func NewLogReader(segmentPath string, startOffset int64) (*LogReader, error) {
	file, err := os.Open(segmentPath)
	if err != nil {
		return nil, err
	}

	// Seek to start offset
	if startOffset > 0 {
		if _, err := file.Seek(startOffset, io.SeekStart); err != nil {
			file.Close()
			return nil, err
		}
	}

	return &LogReader{
		file:       file,
		reader:     bufio.NewReader(file),
		segment:    segmentPath,
		offset:     startOffset,
		lineNumber: 0,
	}, nil
}

// ReadEntry reads the next log entry and returns it with position info
type LogRecord struct {
	Entry      logger.LogEntry
	Offset     int64 // Byte offset AFTER this entry
	LineNumber int64 // Line number of this entry
	Raw        []byte
}

// Read reads the next log entry from the segment
func (lr *LogReader) Read() (*LogRecord, error) {
	line, err := lr.reader.ReadBytes('\n')
	if err != nil {
		if err == io.EOF && len(line) == 0 {
			return nil, io.EOF
		}
		if err != io.EOF {
			return nil, err
		}
	}

	// Update position
	lr.offset += int64(len(line))
	lr.lineNumber++

	// Parse JSON log entry
	var entry logger.LogEntry
	if err := json.Unmarshal(line, &entry); err != nil {
		// Return raw line even if parsing fails
		return &LogRecord{
			Offset:     lr.offset,
			LineNumber: lr.lineNumber,
			Raw:        line,
		}, nil
	}

	return &LogRecord{
		Entry:      entry,
		Offset:     lr.offset,
		LineNumber: lr.lineNumber,
		Raw:        line,
	}, nil
}

// Offset returns the current byte offset
func (lr *LogReader) Offset() int64 {
	return lr.offset
}

// LineNumber returns the current line number
func (lr *LogReader) LineNumber() int64 {
	return lr.lineNumber
}

// Close closes the reader
func (lr *LogReader) Close() error {
	return lr.file.Close()
}
