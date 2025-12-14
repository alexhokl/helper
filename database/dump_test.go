package database

import (
	"testing"
	"time"
)

func TestGetStringValueNil(t *testing.T) {
	var val interface{} = nil
	result := getStringValue(&val)
	if result != "NULL" {
		t.Errorf("getStringValue(nil) = %q, want %q", result, "NULL")
	}
}

func TestGetStringValueBoolTrue(t *testing.T) {
	var val interface{} = true
	result := getStringValue(&val)
	if result != "TRUE" {
		t.Errorf("getStringValue(true) = %q, want %q", result, "TRUE")
	}
}

func TestGetStringValueBoolFalse(t *testing.T) {
	var val interface{} = false
	result := getStringValue(&val)
	if result != "FALSE" {
		t.Errorf("getStringValue(false) = %q, want %q", result, "FALSE")
	}
}

func TestGetStringValueBytes(t *testing.T) {
	var val interface{} = []byte("hello world")
	result := getStringValue(&val)
	if result != "hello world" {
		t.Errorf("getStringValue([]byte) = %q, want %q", result, "hello world")
	}
}

func TestGetStringValueBytesEmpty(t *testing.T) {
	var val interface{} = []byte{}
	result := getStringValue(&val)
	if result != "" {
		t.Errorf("getStringValue([]byte{}) = %q, want empty string", result)
	}
}

func TestGetStringValueTime(t *testing.T) {
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 123000000, time.UTC)
	var val interface{} = testTime
	result := getStringValue(&val)
	expected := "2023-06-15 14:30:45.123"
	if result != expected {
		t.Errorf("getStringValue(time.Time) = %q, want %q", result, expected)
	}
}

func TestGetStringValueTimeNoMilliseconds(t *testing.T) {
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC)
	var val interface{} = testTime
	result := getStringValue(&val)
	expected := "2023-06-15 14:30:45"
	if result != expected {
		t.Errorf("getStringValue(time.Time) = %q, want %q", result, expected)
	}
}

func TestGetStringValueString(t *testing.T) {
	var val interface{} = "test string"
	result := getStringValue(&val)
	if result != "test string" {
		t.Errorf("getStringValue(string) = %q, want %q", result, "test string")
	}
}

func TestGetStringValueInt(t *testing.T) {
	var val interface{} = 42
	result := getStringValue(&val)
	if result != "42" {
		t.Errorf("getStringValue(int) = %q, want %q", result, "42")
	}
}

func TestGetStringValueFloat(t *testing.T) {
	var val interface{} = 3.14
	result := getStringValue(&val)
	if result != "3.14" {
		t.Errorf("getStringValue(float64) = %q, want %q", result, "3.14")
	}
}

func TestGetStringValueNegativeInt(t *testing.T) {
	var val interface{} = -100
	result := getStringValue(&val)
	if result != "-100" {
		t.Errorf("getStringValue(-100) = %q, want %q", result, "-100")
	}
}

func TestGetStringValues(t *testing.T) {
	var val1 interface{} = "hello"
	var val2 interface{} = 42
	var val3 interface{} = true

	row := []interface{}{&val1, &val2, &val3}
	result := getStringValues(row)

	if len(result) != 3 {
		t.Fatalf("getStringValues() returned %d items, want 3", len(result))
	}

	if result[0] != "hello" {
		t.Errorf("result[0] = %q, want %q", result[0], "hello")
	}
	if result[1] != "42" {
		t.Errorf("result[1] = %q, want %q", result[1], "42")
	}
	if result[2] != "TRUE" {
		t.Errorf("result[2] = %q, want %q", result[2], "TRUE")
	}
}

func TestGetStringValuesEmpty(t *testing.T) {
	row := []interface{}{}
	result := getStringValues(row)

	if len(result) != 0 {
		t.Errorf("getStringValues([]) returned %d items, want 0", len(result))
	}
}

func TestGetStringValuesMixedTypes(t *testing.T) {
	var val1 interface{} = nil
	var val2 interface{} = []byte("bytes")
	var val3 interface{} = false
	testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	var val4 interface{} = testTime

	row := []interface{}{&val1, &val2, &val3, &val4}
	result := getStringValues(row)

	if len(result) != 4 {
		t.Fatalf("getStringValues() returned %d items, want 4", len(result))
	}

	if result[0] != "NULL" {
		t.Errorf("result[0] = %q, want %q", result[0], "NULL")
	}
	if result[1] != "bytes" {
		t.Errorf("result[1] = %q, want %q", result[1], "bytes")
	}
	if result[2] != "FALSE" {
		t.Errorf("result[2] = %q, want %q", result[2], "FALSE")
	}
	if result[3] != "2023-01-01 00:00:00" {
		t.Errorf("result[3] = %q, want %q", result[3], "2023-01-01 00:00:00")
	}
}

func TestDumpTablesEmpty(t *testing.T) {
	err := DumpTables([]TableData{})
	if err != nil {
		t.Errorf("DumpTables([]) error: %v", err)
	}
}

func TestDumpTableWithData(t *testing.T) {
	var val1 interface{} = "row1col1"
	var val2 interface{} = 100

	data := &TableData{
		Columns: []string{"Column1", "Column2"},
		Rows: [][]interface{}{
			{&val1, &val2},
		},
	}

	// DumpTable writes to stdout, we just verify it doesn't error
	err := DumpTable(data)
	if err != nil {
		t.Errorf("DumpTable() error: %v", err)
	}
}

func TestDumpTableEmptyRows(t *testing.T) {
	data := &TableData{
		Columns: []string{"Column1", "Column2"},
		Rows:    [][]interface{}{},
	}

	err := DumpTable(data)
	if err != nil {
		t.Errorf("DumpTable() with empty rows error: %v", err)
	}
}

func TestDumpTableEmptyColumns(t *testing.T) {
	data := &TableData{
		Columns: []string{},
		Rows:    [][]interface{}{},
	}

	err := DumpTable(data)
	if err != nil {
		t.Errorf("DumpTable() with empty columns error: %v", err)
	}
}

func TestDumpTablesMultipleTables(t *testing.T) {
	var val1 interface{} = "data1"
	var val2 interface{} = "data2"

	tables := []TableData{
		{
			Columns: []string{"Col1"},
			Rows:    [][]interface{}{{&val1}},
		},
		{
			Columns: []string{"Col2"},
			Rows:    [][]interface{}{{&val2}},
		},
	}

	err := DumpTables(tables)
	if err != nil {
		t.Errorf("DumpTables() error: %v", err)
	}
}

// Benchmark tests
func BenchmarkGetStringValue(b *testing.B) {
	var val interface{} = "benchmark string"
	for i := 0; i < b.N; i++ {
		_ = getStringValue(&val)
	}
}

func BenchmarkGetStringValueTime(b *testing.B) {
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 123000000, time.UTC)
	var val interface{} = testTime
	for i := 0; i < b.N; i++ {
		_ = getStringValue(&val)
	}
}

func BenchmarkGetStringValues(b *testing.B) {
	var val1 interface{} = "hello"
	var val2 interface{} = 42
	var val3 interface{} = true
	row := []interface{}{&val1, &val2, &val3}

	for i := 0; i < b.N; i++ {
		_ = getStringValues(row)
	}
}
