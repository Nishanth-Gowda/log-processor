package logger

import (
	"encoding/json"
	"testing"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
	"github.com/minio/simdjson-go"
)

// Sample log entries for benchmarking
var sampleLogs = []LogEntry{
	{Timestamp: "2026-01-01T16:38:14.328717Z", Level: INFO, Service: "payment-service", Message: "File uploaded", RequestID: "req-53b55783", Duration: 1644},
	{Timestamp: "2026-01-01T16:38:14.828588Z", Level: ERROR, Service: "user-service", Message: "Service timeout", RequestID: "req-3b33af3e", UserID: "user-9956"},
	{Timestamp: "2026-01-01T16:38:15.328578Z", Level: INFO, Service: "payment-service", Message: "Session created", RequestID: "req-fb6370dc", Duration: 2821},
	{Timestamp: "2026-01-01T16:38:15.828587Z", Level: WARNING, Service: "auth-service", Message: "High memory usage", RequestID: "req-be8abace", UserID: "user-4992", Duration: 2557},
	{Timestamp: "2026-01-01T16:38:16.32858Z", Level: INFO, Service: "api-gateway", Message: "User logged in", RequestID: "req-641efe00", UserID: "user-1432", Duration: 1035},
}

// Pre-serialized JSON for parsing benchmarks
var sampleJSONBytes [][]byte

func init() {
	sampleJSONBytes = make([][]byte, len(sampleLogs))
	for i, log := range sampleLogs {
		data, _ := json.Marshal(log)
		sampleJSONBytes[i] = data
	}
}

// ============================================================================
// ENCODING BENCHMARKS (struct -> JSON)
// ============================================================================

// BenchmarkEncodingJSON benchmarks standard encoding/json Marshal
func BenchmarkEncodingJSON_Marshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, log := range sampleLogs {
			_, err := json.Marshal(log)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkEncodingJSON_MarshalSingle benchmarks marshaling a single log entry
func BenchmarkEncodingJSON_MarshalSingle(b *testing.B) {
	log := sampleLogs[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(log)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// DECODING/PARSING BENCHMARKS (JSON -> struct/data)
// ============================================================================

// BenchmarkEncodingJSON_Unmarshal benchmarks standard encoding/json Unmarshal
func BenchmarkEncodingJSON_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, data := range sampleJSONBytes {
			var log LogEntry
			err := json.Unmarshal(data, &log)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkEncodingJSON_UnmarshalSingle benchmarks unmarshaling a single log entry
func BenchmarkEncodingJSON_UnmarshalSingle(b *testing.B) {
	data := sampleJSONBytes[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var log LogEntry
		err := json.Unmarshal(data, &log)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSimdJSON_Parse benchmarks simdjson-go parsing
func BenchmarkSimdJSON_Parse(b *testing.B) {
	if !simdjson.SupportedCPU() {
		b.Skip("CPU does not support simdjson")
	}

	// Combine all JSON into NDJSON format for simdjson
	var ndjson []byte
	for _, data := range sampleJSONBytes {
		ndjson = append(ndjson, data...)
		ndjson = append(ndjson, '\n')
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pj, err := simdjson.Parse(ndjson, nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = pj
	}
}

// BenchmarkSimdJSON_ParseSingle benchmarks simdjson-go parsing a single entry
func BenchmarkSimdJSON_ParseSingle(b *testing.B) {
	if !simdjson.SupportedCPU() {
		b.Skip("CPU does not support simdjson")
	}

	data := sampleJSONBytes[0]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pj, err := simdjson.Parse(data, nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = pj
	}
}

// BenchmarkSimdJSON_ParseAndExtract benchmarks simdjson-go parse + field extraction
func BenchmarkSimdJSON_ParseAndExtract(b *testing.B) {
	if !simdjson.SupportedCPU() {
		b.Skip("CPU does not support simdjson")
	}

	data := sampleJSONBytes[0]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pj, err := simdjson.Parse(data, nil)
		if err != nil {
			b.Fatal(err)
		}

		// Extract fields using iterator
		_ = pj.ForEach(func(iter simdjson.Iter) error {
			// Find and extract specific fields
			if elem, err := iter.FindElement(nil, "service"); err == nil {
				_, _ = elem.Iter.StringBytes()
			}
			if elem, err := iter.FindElement(nil, "level"); err == nil {
				_, _ = elem.Iter.StringBytes()
			}
			if elem, err := iter.FindElement(nil, "message"); err == nil {
				_, _ = elem.Iter.StringBytes()
			}
			return nil
		})
	}
}

// BenchmarkEncodingJSON_UnmarshalToInterface benchmarks unmarshaling to interface{}
func BenchmarkEncodingJSON_UnmarshalToInterface(b *testing.B) {
	data := sampleJSONBytes[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result interface{}
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSimdJSON_ParseToInterface benchmarks simdjson parsing to interface{}
func BenchmarkSimdJSON_ParseToInterface(b *testing.B) {
	if !simdjson.SupportedCPU() {
		b.Skip("CPU does not support simdjson")
	}

	data := sampleJSONBytes[0]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pj, err := simdjson.Parse(data, nil)
		if err != nil {
			b.Fatal(err)
		}

		_ = pj.ForEach(func(iter simdjson.Iter) error {
			_, _ = iter.Interface()
			return nil
		})
	}
}

// ============================================================================
// NDJSON PARSING BENCHMARKS (multiple records)
// ============================================================================

// BenchmarkEncodingJSON_UnmarshalNDJSON benchmarks parsing multiple JSON lines
func BenchmarkEncodingJSON_UnmarshalNDJSON(b *testing.B) {
	// Create NDJSON data
	var ndjson []byte
	for _, data := range sampleJSONBytes {
		ndjson = append(ndjson, data...)
		ndjson = append(ndjson, '\n')
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Parse line by line (simulating NDJSON parsing)
		for _, data := range sampleJSONBytes {
			var log LogEntry
			_ = json.Unmarshal(data, &log)
		}
	}
}

// BenchmarkSimdJSON_ParseNDJSON benchmarks simdjson NDJSON parsing
func BenchmarkSimdJSON_ParseNDJSON(b *testing.B) {
	if !simdjson.SupportedCPU() {
		b.Skip("CPU does not support simdjson")
	}

	// Create NDJSON data
	var ndjson []byte
	for _, data := range sampleJSONBytes {
		ndjson = append(ndjson, data...)
		ndjson = append(ndjson, '\n')
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pj, err := simdjson.ParseND(ndjson, nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = pj
	}
}

// ============================================================================
// SONIC BENCHMARKS (ARM64 compatible high-performance JSON)
// ============================================================================

// BenchmarkSonic_Marshal benchmarks Sonic Marshal (all logs)
func BenchmarkSonic_Marshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, log := range sampleLogs {
			_, err := sonic.Marshal(log)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkSonic_MarshalSingle benchmarks Sonic Marshal (single log)
func BenchmarkSonic_MarshalSingle(b *testing.B) {
	log := sampleLogs[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sonic.Marshal(log)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSonic_Unmarshal benchmarks Sonic Unmarshal (all logs)
func BenchmarkSonic_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, data := range sampleJSONBytes {
			var log LogEntry
			err := sonic.Unmarshal(data, &log)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkSonic_UnmarshalSingle benchmarks Sonic Unmarshal (single log)
func BenchmarkSonic_UnmarshalSingle(b *testing.B) {
	data := sampleJSONBytes[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var log LogEntry
		err := sonic.Unmarshal(data, &log)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSonic_UnmarshalToInterface benchmarks Sonic unmarshal to interface{}
func BenchmarkSonic_UnmarshalToInterface(b *testing.B) {
	data := sampleJSONBytes[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result interface{}
		err := sonic.Unmarshal(data, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSonic_UnmarshalNDJSON benchmarks Sonic parsing multiple JSON lines
func BenchmarkSonic_UnmarshalNDJSON(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range sampleJSONBytes {
			var log LogEntry
			_ = sonic.Unmarshal(data, &log)
		}
	}
}

// ============================================================================
// GOCCY/GO-JSON BENCHMARKS (balanced performance)
// ============================================================================

// BenchmarkGoJSON_Marshal benchmarks go-json Marshal (all logs)
func BenchmarkGoJSON_Marshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, log := range sampleLogs {
			_, err := gojson.Marshal(log)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkGoJSON_MarshalSingle benchmarks go-json Marshal (single log)
func BenchmarkGoJSON_MarshalSingle(b *testing.B) {
	log := sampleLogs[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gojson.Marshal(log)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGoJSON_Unmarshal benchmarks go-json Unmarshal (all logs)
func BenchmarkGoJSON_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, data := range sampleJSONBytes {
			var log LogEntry
			err := gojson.Unmarshal(data, &log)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkGoJSON_UnmarshalSingle benchmarks go-json Unmarshal (single log)
func BenchmarkGoJSON_UnmarshalSingle(b *testing.B) {
	data := sampleJSONBytes[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var log LogEntry
		err := gojson.Unmarshal(data, &log)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// JSONITER BENCHMARKS (json-iterator/go)
// ============================================================================

var jsoniterAPI = jsoniter.ConfigCompatibleWithStandardLibrary

// BenchmarkJsoniter_Marshal benchmarks jsoniter Marshal (all logs)
func BenchmarkJsoniter_Marshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, log := range sampleLogs {
			_, err := jsoniterAPI.Marshal(log)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkJsoniter_MarshalSingle benchmarks jsoniter Marshal (single log)
func BenchmarkJsoniter_MarshalSingle(b *testing.B) {
	log := sampleLogs[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jsoniterAPI.Marshal(log)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJsoniter_Unmarshal benchmarks jsoniter Unmarshal (all logs)
func BenchmarkJsoniter_Unmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, data := range sampleJSONBytes {
			var log LogEntry
			err := jsoniterAPI.Unmarshal(data, &log)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkJsoniter_UnmarshalSingle benchmarks jsoniter Unmarshal (single log)
func BenchmarkJsoniter_UnmarshalSingle(b *testing.B) {
	data := sampleJSONBytes[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var log LogEntry
		err := jsoniterAPI.Unmarshal(data, &log)
		if err != nil {
			b.Fatal(err)
		}
	}
}
