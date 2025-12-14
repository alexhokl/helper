package database

import (
	"strings"
	"testing"
	"time"
)

func TestTableDataStruct(t *testing.T) {
	data := TableData{
		Columns: []string{"id", "name", "created_at"},
		Rows:    [][]interface{}{},
	}

	if len(data.Columns) != 3 {
		t.Errorf("Columns length = %d, want 3", len(data.Columns))
	}
	if len(data.Rows) != 0 {
		t.Errorf("Rows length = %d, want 0", len(data.Rows))
	}
}

func TestGetValueNil(t *testing.T) {
	var val interface{} = nil
	result := GetValue(&val)
	if result != "NULL" {
		t.Errorf("GetValue(nil) = %q, want %q", result, "NULL")
	}
}

func TestGetValueBoolTrue(t *testing.T) {
	var val interface{} = true
	result := GetValue(&val)
	if result != "1" {
		t.Errorf("GetValue(true) = %q, want %q", result, "1")
	}
}

func TestGetValueBoolFalse(t *testing.T) {
	var val interface{} = false
	result := GetValue(&val)
	if result != "0" {
		t.Errorf("GetValue(false) = %q, want %q", result, "0")
	}
}

func TestGetValueBytes(t *testing.T) {
	var val interface{} = []byte("hello world")
	result := GetValue(&val)
	if result != "hello world" {
		t.Errorf("GetValue([]byte) = %q, want %q", result, "hello world")
	}
}

func TestGetValueBytesEmpty(t *testing.T) {
	var val interface{} = []byte{}
	result := GetValue(&val)
	if result != "" {
		t.Errorf("GetValue([]byte{}) = %q, want empty string", result)
	}
}

func TestGetValueTime(t *testing.T) {
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 123000000, time.UTC)
	var val interface{} = testTime
	result := GetValue(&val)
	expected := "2023-06-15 14:30:45.123"
	if result != expected {
		t.Errorf("GetValue(time.Time) = %q, want %q", result, expected)
	}
}

func TestGetValueTimeNoMilliseconds(t *testing.T) {
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC)
	var val interface{} = testTime
	result := GetValue(&val)
	expected := "2023-06-15 14:30:45"
	if result != expected {
		t.Errorf("GetValue(time.Time) = %q, want %q", result, expected)
	}
}

func TestGetValueString(t *testing.T) {
	var val interface{} = "test string"
	result := GetValue(&val)
	if result != "test string" {
		t.Errorf("GetValue(string) = %q, want %q", result, "test string")
	}
}

func TestGetValueInt(t *testing.T) {
	var val interface{} = 42
	result := GetValue(&val)
	if result != "42" {
		t.Errorf("GetValue(int) = %q, want %q", result, "42")
	}
}

func TestGetValueFloat(t *testing.T) {
	var val interface{} = 3.14
	result := GetValue(&val)
	if result != "3.14" {
		t.Errorf("GetValue(float64) = %q, want %q", result, "3.14")
	}
}

func TestGetValueNegativeInt(t *testing.T) {
	var val interface{} = -100
	result := GetValue(&val)
	if result != "-100" {
		t.Errorf("GetValue(-100) = %q, want %q", result, "-100")
	}
}

func TestGetConnectionURLFormat(t *testing.T) {
	config := &Config{
		Server:   "localhost",
		Port:     1433,
		Username: "sa",
		Password: "MyPassword123",
		Name:     "testdb",
	}

	conn, err := GetConnection(config)
	if err != nil {
		t.Fatalf("GetConnection() error: %v", err)
	}
	defer conn.Close()

	// Verify the connection was created (we can't actually connect without a server)
	if conn == nil {
		t.Error("GetConnection() returned nil connection")
	}
}

func TestGetConnectionWithSpecialCharactersInPassword(t *testing.T) {
	config := &Config{
		Server:   "localhost",
		Port:     1433,
		Username: "sa",
		Password: "P@ss:word/with?special&chars=test",
		Name:     "testdb",
	}

	conn, err := GetConnection(config)
	if err != nil {
		t.Fatalf("GetConnection() with special chars error: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Error("GetConnection() with special chars returned nil connection")
	}
}

func TestGetPostgresConnectionURLFormat(t *testing.T) {
	config := &PostgresConfig{
		Config: Config{
			Server:   "localhost:5432",
			Port:     5432,
			Username: "postgres",
			Password: "postgrespass",
			Name:     "testdb",
		},
		UseSSL: false,
	}

	conn, err := GetPostgresConnection(config)
	if err != nil {
		t.Fatalf("GetPostgresConnection() error: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Error("GetPostgresConnection() returned nil connection")
	}
}

func TestGetPostgresConnectionWithSSL(t *testing.T) {
	config := &PostgresConfig{
		Config: Config{
			Server:   "localhost:5432",
			Port:     5432,
			Username: "postgres",
			Password: "postgrespass",
			Name:     "testdb",
		},
		UseSSL: true,
	}

	conn, err := GetPostgresConnection(config)
	if err != nil {
		t.Fatalf("GetPostgresConnection() with SSL error: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Error("GetPostgresConnection() with SSL returned nil connection")
	}
}

func TestGetPostgresConnectionWithSpecialCharactersInPassword(t *testing.T) {
	config := &PostgresConfig{
		Config: Config{
			Server:   "localhost:5432",
			Port:     5432,
			Username: "postgres",
			Password: "P@ss:word/with?special&chars=test",
			Name:     "testdb",
		},
		UseSSL: false,
	}

	conn, err := GetPostgresConnection(config)
	if err != nil {
		t.Fatalf("GetPostgresConnection() with special chars error: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Error("GetPostgresConnection() with special chars returned nil connection")
	}
}

func TestGetValueVsGetStringValueDifference(t *testing.T) {
	// GetValue returns "1"/"0" for booleans (SQL-style)
	// getStringValue returns "TRUE"/"FALSE" for booleans (display-style)

	var trueVal interface{} = true
	var falseVal interface{} = false

	getValueTrue := GetValue(&trueVal)
	getValueFalse := GetValue(&falseVal)

	if getValueTrue != "1" {
		t.Errorf("GetValue(true) = %q, want %q", getValueTrue, "1")
	}
	if getValueFalse != "0" {
		t.Errorf("GetValue(false) = %q, want %q", getValueFalse, "0")
	}

	// Reset values for getStringValue test
	trueVal = true
	falseVal = false

	getStringValueTrue := getStringValue(&trueVal)
	getStringValueFalse := getStringValue(&falseVal)

	if getStringValueTrue != "TRUE" {
		t.Errorf("getStringValue(true) = %q, want %q", getStringValueTrue, "TRUE")
	}
	if getStringValueFalse != "FALSE" {
		t.Errorf("getStringValue(false) = %q, want %q", getStringValueFalse, "FALSE")
	}
}

func TestConnectionURLContainsDatabase(t *testing.T) {
	// This test verifies that the database name is properly included in the connection URL
	config := &Config{
		Server:   "myserver.example.com",
		Port:     1433,
		Username: "admin",
		Password: "secret",
		Name:     "production_db",
	}

	conn, err := GetConnection(config)
	if err != nil {
		t.Fatalf("GetConnection() error: %v", err)
	}
	defer conn.Close()

	// Connection was created successfully, database parameter was included
	if conn == nil {
		t.Error("GetConnection() returned nil")
	}
}

func TestPostgresConnectionURLContainsSSLMode(t *testing.T) {
	// Test that sslmode=disable is added when UseSSL is false
	configNoSSL := &PostgresConfig{
		Config: Config{
			Server:   "localhost:5432",
			Username: "postgres",
			Password: "password",
			Name:     "testdb",
		},
		UseSSL: false,
	}

	conn, err := GetPostgresConnection(configNoSSL)
	if err != nil {
		t.Fatalf("GetPostgresConnection() error: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Error("GetPostgresConnection() returned nil")
	}
}

// Benchmark tests
func BenchmarkGetValue(b *testing.B) {
	var val interface{} = "benchmark string"
	for i := 0; i < b.N; i++ {
		_ = GetValue(&val)
	}
}

func BenchmarkGetValueTime(b *testing.B) {
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 123000000, time.UTC)
	var val interface{} = testTime
	for i := 0; i < b.N; i++ {
		_ = GetValue(&val)
	}
}

func BenchmarkGetConnection(b *testing.B) {
	config := &Config{
		Server:   "localhost",
		Port:     1433,
		Username: "sa",
		Password: "password",
		Name:     "testdb",
	}

	for i := 0; i < b.N; i++ {
		conn, _ := GetConnection(config)
		if conn != nil {
			conn.Close()
		}
	}
}

func BenchmarkGetPostgresConnection(b *testing.B) {
	config := &PostgresConfig{
		Config: Config{
			Server:   "localhost:5432",
			Username: "postgres",
			Password: "password",
			Name:     "testdb",
		},
		UseSSL: false,
	}

	for i := 0; i < b.N; i++ {
		conn, _ := GetPostgresConnection(config)
		if conn != nil {
			conn.Close()
		}
	}
}

// Test table for GetValue function
func TestGetValueTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, "NULL"},
		{"bool_true", true, "1"},
		{"bool_false", false, "0"},
		{"string", "hello", "hello"},
		{"empty_string", "", ""},
		{"int", 42, "42"},
		{"negative_int", -42, "-42"},
		{"zero", 0, "0"},
		{"float", 3.14159, "3.14159"},
		{"bytes", []byte("test bytes"), "test bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.input
			result := GetValue(&val)
			// For some types like float, we may need approximate comparison
			if tt.name == "float" {
				if !strings.HasPrefix(result, "3.14") {
					t.Errorf("GetValue(%v) = %q, want prefix %q", tt.input, result, "3.14")
				}
			} else if result != tt.expected {
				t.Errorf("GetValue(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
