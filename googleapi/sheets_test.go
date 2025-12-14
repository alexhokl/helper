package googleapi

import (
	"testing"
	"time"

	sheets "google.golang.org/api/sheets/v4"
)

func TestGetValueNil(t *testing.T) {
	var val interface{} = nil
	result := getValue(&val)
	if result != "NULL" {
		t.Errorf("getValue(nil) = %q, want %q", result, "NULL")
	}
}

func TestGetValueBoolTrue(t *testing.T) {
	var val interface{} = true
	result := getValue(&val)
	if result != "1" {
		t.Errorf("getValue(true) = %q, want %q", result, "1")
	}
}

func TestGetValueBoolFalse(t *testing.T) {
	var val interface{} = false
	result := getValue(&val)
	if result != "0" {
		t.Errorf("getValue(false) = %q, want %q", result, "0")
	}
}

func TestGetValueBytes(t *testing.T) {
	var val interface{} = []byte("hello world")
	result := getValue(&val)
	if result != "hello world" {
		t.Errorf("getValue([]byte) = %q, want %q", result, "hello world")
	}
}

func TestGetValueBytesEmpty(t *testing.T) {
	var val interface{} = []byte{}
	result := getValue(&val)
	if result != "" {
		t.Errorf("getValue([]byte{}) = %q, want empty string", result)
	}
}

func TestGetValueTime(t *testing.T) {
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 123000000, time.UTC)
	var val interface{} = testTime
	result := getValue(&val)
	expected := "2023-06-15 14:30:45.123"
	if result != expected {
		t.Errorf("getValue(time.Time) = %q, want %q", result, expected)
	}
}

func TestGetValueTimeNoMilliseconds(t *testing.T) {
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC)
	var val interface{} = testTime
	result := getValue(&val)
	expected := "2023-06-15 14:30:45"
	if result != expected {
		t.Errorf("getValue(time.Time) = %q, want %q", result, expected)
	}
}

func TestGetValueString(t *testing.T) {
	var val interface{} = "test string"
	result := getValue(&val)
	if result != "test string" {
		t.Errorf("getValue(string) = %q, want %q", result, "test string")
	}
}

func TestGetValueInt(t *testing.T) {
	var val interface{} = 42
	result := getValue(&val)
	if result != "42" {
		t.Errorf("getValue(int) = %q, want %q", result, "42")
	}
}

func TestGetValueFloat(t *testing.T) {
	var val interface{} = 3.14
	result := getValue(&val)
	if result != "3.14" {
		t.Errorf("getValue(float64) = %q, want %q", result, "3.14")
	}
}

func TestNewColumnHeadersValueRangeEmpty(t *testing.T) {
	columns := []string{}
	result := newColumnHeadersValueRange(columns)

	if result == nil {
		t.Fatal("newColumnHeadersValueRange() returned nil")
	}

	if len(result.Values) != 1 {
		t.Errorf("Values length = %d, want 1", len(result.Values))
	}

	if len(result.Values[0]) != 0 {
		t.Errorf("Values[0] length = %d, want 0", len(result.Values[0]))
	}
}

func TestNewColumnHeadersValueRangeSingle(t *testing.T) {
	columns := []string{"Column1"}
	result := newColumnHeadersValueRange(columns)

	if len(result.Values[0]) != 1 {
		t.Errorf("Values[0] length = %d, want 1", len(result.Values[0]))
	}

	if result.Values[0][0] != "Column1" {
		t.Errorf("Values[0][0] = %v, want %q", result.Values[0][0], "Column1")
	}
}

func TestNewColumnHeadersValueRangeMultiple(t *testing.T) {
	columns := []string{"ID", "Name", "Email", "CreatedAt"}
	result := newColumnHeadersValueRange(columns)

	if len(result.Values[0]) != 4 {
		t.Errorf("Values[0] length = %d, want 4", len(result.Values[0]))
	}

	for i, col := range columns {
		if result.Values[0][i] != col {
			t.Errorf("Values[0][%d] = %v, want %q", i, result.Values[0][i], col)
		}
	}
}

func TestNewSpreadSheet(t *testing.T) {
	documentName := "Test Document"
	result := newSpreadSheet(documentName)

	if result == nil {
		t.Fatal("newSpreadSheet() returned nil")
	}

	if result.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if result.Properties.Title != documentName {
		t.Errorf("Title = %q, want %q", result.Properties.Title, documentName)
	}
}

func TestNewSpreadSheetEmptyName(t *testing.T) {
	documentName := ""
	result := newSpreadSheet(documentName)

	if result.Properties.Title != "" {
		t.Errorf("Title = %q, want empty string", result.Properties.Title)
	}
}

func TestNewSheetProperties(t *testing.T) {
	sheetName := "Sheet1"
	var val1 interface{} = "data1"
	var val2 interface{} = "data2"
	rows := [][]interface{}{
		{&val1},
		{&val2},
	}
	columns := []string{"Col1", "Col2", "Col3"}

	result := newSheetProperties(sheetName, rows, columns)

	if result == nil {
		t.Fatal("newSheetProperties() returned nil")
	}

	if result.Title != sheetName {
		t.Errorf("Title = %q, want %q", result.Title, sheetName)
	}

	if result.GridProperties == nil {
		t.Fatal("GridProperties should not be nil")
	}

	if result.GridProperties.RowCount != 2 {
		t.Errorf("RowCount = %d, want 2", result.GridProperties.RowCount)
	}

	if result.GridProperties.ColumnCount != 3 {
		t.Errorf("ColumnCount = %d, want 3", result.GridProperties.ColumnCount)
	}

	if result.GridProperties.FrozenColumnCount != 1 {
		t.Errorf("FrozenColumnCount = %d, want 1", result.GridProperties.FrozenColumnCount)
	}
}

func TestNewSheetPropertiesEmpty(t *testing.T) {
	sheetName := "EmptySheet"
	rows := [][]interface{}{}
	columns := []string{}

	result := newSheetProperties(sheetName, rows, columns)

	if result.GridProperties.RowCount != 0 {
		t.Errorf("RowCount = %d, want 0", result.GridProperties.RowCount)
	}

	if result.GridProperties.ColumnCount != 0 {
		t.Errorf("ColumnCount = %d, want 0", result.GridProperties.ColumnCount)
	}
}

func TestNewNumberFormatRequest(t *testing.T) {
	sheetId := int64(123)
	columnNumber := int64(5)

	result := newNumberFormatRequest(sheetId, columnNumber)

	if result == nil {
		t.Fatal("newNumberFormatRequest() returned nil")
	}

	if result.RepeatCell == nil {
		t.Fatal("RepeatCell should not be nil")
	}

	if result.RepeatCell.Range == nil {
		t.Fatal("Range should not be nil")
	}

	if result.RepeatCell.Range.SheetId != sheetId {
		t.Errorf("SheetId = %d, want %d", result.RepeatCell.Range.SheetId, sheetId)
	}

	if result.RepeatCell.Range.StartColumnIndex != columnNumber {
		t.Errorf("StartColumnIndex = %d, want %d", result.RepeatCell.Range.StartColumnIndex, columnNumber)
	}

	if result.RepeatCell.Range.EndColumnIndex != columnNumber+1 {
		t.Errorf("EndColumnIndex = %d, want %d", result.RepeatCell.Range.EndColumnIndex, columnNumber+1)
	}

	if result.RepeatCell.Fields != "userEnteredFormat.numberFormat" {
		t.Errorf("Fields = %q, want %q", result.RepeatCell.Fields, "userEnteredFormat.numberFormat")
	}
}

func TestNewDateFormatRequest(t *testing.T) {
	sheetId := int64(123)
	columnNumber := int64(2)
	format := "yyyy-MM-dd"

	result := newDateFormatRequest(sheetId, columnNumber, format)

	if result.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Type != "DATE" {
		t.Errorf("Type = %q, want %q", result.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Type, "DATE")
	}

	if result.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Pattern != format {
		t.Errorf("Pattern = %q, want %q", result.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Pattern, format)
	}
}

func TestNewMoneyFormatRequest(t *testing.T) {
	sheetId := int64(123)
	columnNumber := int64(3)

	result := newMoneyFormatRequest(sheetId, columnNumber)

	if result.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Type != "NUMBER" {
		t.Errorf("Type = %q, want %q", result.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Type, "NUMBER")
	}

	expectedPattern := "#,##0.00;(#,##0.00)"
	if result.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Pattern != expectedPattern {
		t.Errorf("Pattern = %q, want %q", result.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Pattern, expectedPattern)
	}
}

func TestNewCreateSheetRequestFirstSheet(t *testing.T) {
	sheetNo := 0
	sheetName := "First Sheet"
	rows := [][]interface{}{}
	columns := []string{"Col1"}

	result, err := newCreateSheetRequest(sheetNo, sheetName, rows, columns)
	if err != nil {
		t.Fatalf("newCreateSheetRequest() error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if len(result.Requests) != 1 {
		t.Fatalf("Requests length = %d, want 1", len(result.Requests))
	}

	// First sheet (index 0) should use UpdateSheetProperties
	if result.Requests[0].UpdateSheetProperties == nil {
		t.Error("UpdateSheetProperties should not be nil for first sheet")
	}

	if result.Requests[0].AddSheet != nil {
		t.Error("AddSheet should be nil for first sheet")
	}
}

func TestNewCreateSheetRequestSubsequentSheet(t *testing.T) {
	sheetNo := 1
	sheetName := "Second Sheet"
	rows := [][]interface{}{}
	columns := []string{"Col1"}

	result, err := newCreateSheetRequest(sheetNo, sheetName, rows, columns)
	if err != nil {
		t.Fatalf("newCreateSheetRequest() error: %v", err)
	}

	// Subsequent sheets should use AddSheet
	if result.Requests[0].AddSheet == nil {
		t.Error("AddSheet should not be nil for subsequent sheets")
	}

	if result.Requests[0].UpdateSheetProperties != nil {
		t.Error("UpdateSheetProperties should be nil for subsequent sheets")
	}
}

func TestNewAddSheetRequest(t *testing.T) {
	sheetName := "New Sheet"
	rows := [][]interface{}{}
	columns := []string{"A", "B"}

	result := newAddSheetRequest(sheetName, rows, columns)

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if result.Properties.Title != sheetName {
		t.Errorf("Title = %q, want %q", result.Properties.Title, sheetName)
	}
}

func TestNewUpdateSheetRequest(t *testing.T) {
	sheetName := "Updated Sheet"
	rows := [][]interface{}{}
	columns := []string{"X", "Y", "Z"}

	result := newUpdateSheetRequest(sheetName, rows, columns)

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if result.Properties.Title != sheetName {
		t.Errorf("Title = %q, want %q", result.Properties.Title, sheetName)
	}

	expectedFields := "Title,GridProperties.RowCount,GridProperties.ColumnCount,GridProperties.FrozenRowCount"
	if result.Fields != expectedFields {
		t.Errorf("Fields = %q, want %q", result.Fields, expectedFields)
	}
}

func TestNewRowsValueRangeEmpty(t *testing.T) {
	rows := [][]interface{}{}
	result := newRowsValueRange(rows)

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if len(result.Values) != 0 {
		t.Errorf("Values length = %d, want 0", len(result.Values))
	}
}

func TestNewColumnFormatRequestNil(t *testing.T) {
	result := newColumnFormatRequest(123, nil)

	if result != nil {
		t.Error("newColumnFormatRequest(nil) should return nil")
	}
}

func TestUpdateColumnStylesNil(t *testing.T) {
	// UpdateColumnStyles with nil columns should return nil (no-op)
	err := UpdateColumnStyles(nil, &sheets.Spreadsheet{SpreadsheetId: "test"}, 0, nil)
	if err != nil {
		t.Errorf("UpdateColumnStyles() with nil columns error: %v", err)
	}
}

// mockColumnFormatConfig implements ColumnFormatConfig for testing
type mockColumnFormatConfig struct {
	index    int
	dataType string
	format   string
}

func (m *mockColumnFormatConfig) ColumnIndex() int {
	return m.index
}

func (m *mockColumnFormatConfig) ColumnDataType() string {
	return m.dataType
}

func (m *mockColumnFormatConfig) ColumnFormat() string {
	return m.format
}

func TestNewColumnFormatRequestDate(t *testing.T) {
	columns := []ColumnFormatConfig{
		&mockColumnFormatConfig{
			index:    0,
			dataType: "date",
			format:   "yyyy-MM-dd",
		},
	}

	result := newColumnFormatRequest(123, columns)

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if len(result.Requests) != 1 {
		t.Fatalf("Requests length = %d, want 1", len(result.Requests))
	}

	if result.Requests[0].RepeatCell.Cell.UserEnteredFormat.NumberFormat.Type != "DATE" {
		t.Error("Type should be DATE")
	}
}

func TestNewColumnFormatRequestMoney(t *testing.T) {
	columns := []ColumnFormatConfig{
		&mockColumnFormatConfig{
			index:    1,
			dataType: "money",
			format:   "",
		},
	}

	result := newColumnFormatRequest(123, columns)

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Requests[0].RepeatCell.Cell.UserEnteredFormat.NumberFormat.Type != "NUMBER" {
		t.Error("Type should be NUMBER")
	}
}

func TestNewColumnFormatRequestUnknownTypePanics(t *testing.T) {
	columns := []ColumnFormatConfig{
		&mockColumnFormatConfig{
			index:    0,
			dataType: "unknown",
			format:   "",
		},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("newColumnFormatRequest() with unknown type should panic")
		}
	}()

	newColumnFormatRequest(123, columns)
}

func TestNewColumnFormatRequestMultipleColumns(t *testing.T) {
	columns := []ColumnFormatConfig{
		&mockColumnFormatConfig{index: 0, dataType: "date", format: "yyyy-MM-dd"},
		&mockColumnFormatConfig{index: 1, dataType: "money", format: ""},
		&mockColumnFormatConfig{index: 2, dataType: "date", format: "MM/dd/yyyy"},
	}

	result := newColumnFormatRequest(123, columns)

	if len(result.Requests) != 3 {
		t.Errorf("Requests length = %d, want 3", len(result.Requests))
	}
}

// Benchmark tests
func BenchmarkGetValue(b *testing.B) {
	var val interface{} = "benchmark string"
	for i := 0; i < b.N; i++ {
		_ = getValue(&val)
	}
}

func BenchmarkGetValueTime(b *testing.B) {
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 123000000, time.UTC)
	var val interface{} = testTime
	for i := 0; i < b.N; i++ {
		_ = getValue(&val)
	}
}

func BenchmarkNewColumnHeadersValueRange(b *testing.B) {
	columns := []string{"ID", "Name", "Email", "Phone", "Address", "City", "State", "Zip"}
	for i := 0; i < b.N; i++ {
		_ = newColumnHeadersValueRange(columns)
	}
}

func BenchmarkNewSpreadSheet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = newSpreadSheet("Benchmark Document")
	}
}
