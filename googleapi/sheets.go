package googleapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

type ColumnFormatConfig interface {
	ColumnIndex() int
	ColumnDataType() string
	ColumnFormat() string
}

func NewSpreadsheetService(ctx context.Context, client *http.Client) (*sheets.Service, error) {
	return sheets.NewService(
		ctx,
		option.WithHTTPClient(client),
	)
}

func CreateSpreadSheet(srv *sheets.Service, documentName string) (*sheets.Spreadsheet, error) {
	s := newSpreadSheet(documentName)
	createdSpreadSheet, err := srv.Spreadsheets.Create(s).Do()
	if err != nil {
		return nil, err
	}
	return createdSpreadSheet, nil
}

func CreateSheet(service *sheets.Service, document *sheets.Spreadsheet, sheetNo int, sheetName string, rows [][]interface{}, columns []string) (int64, error) {
	request, errRequest := newCreateSheetRequest(
		sheetNo,
		sheetName,
		rows,
		columns,
	)
	if errRequest != nil {
		return -1, errRequest
	}

	response, err := service.Spreadsheets.BatchUpdate(document.SpreadsheetId, request).Do()
	if err != nil {
		return -1, err
	}
	if response.Replies[0].AddSheet != nil {
		return response.Replies[0].AddSheet.Properties.SheetId, nil
	}
	return 0, nil
}

func UpdateColumnHeaders(service *sheets.Service, document *sheets.Spreadsheet, sheetName string, columns []string) error {
	valueRange := newColumnHeadersValueRange(columns)
	_, err := service.Spreadsheets.Values.Update(
		document.SpreadsheetId,
		fmt.Sprintf("%s!A1", sheetName),
		valueRange,
	).ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		return err
	}
	return nil
}

func UpdateRows(service *sheets.Service, document *sheets.Spreadsheet, sheetName string, rows [][]interface{}) error {
	values := newRowsValueRange(rows)
	_, err := service.Spreadsheets.Values.Update(
		document.SpreadsheetId,
		fmt.Sprintf("%s!A2", sheetName),
		values,
	).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}
	return nil
}

func UpdateColumnStyles(service *sheets.Service, document *sheets.Spreadsheet, sheetId int64, columns []ColumnFormatConfig) error {
	if columns == nil {
		return nil
	}
	request := newColumnFormatRequest(sheetId, columns)
	_, err := service.Spreadsheets.BatchUpdate(document.SpreadsheetId, request).Do()
	if err != nil {
		return err
	}
	return nil
}

func newRowsValueRange(rows [][]interface{}) *sheets.ValueRange {
	values := &sheets.ValueRange{}
	values.Values = make([][]interface{}, len(rows))
	for i, r := range rows {
		for _, cell := range r {
			values.Values[i] = append(
				values.Values[i],
				getValue(cell.(*interface{})),
			)
		}
	}
	return values
}

func getValue(pval *interface{}) string {
	switch v := (*pval).(type) {
	case nil:
		return "NULL"
	case bool:
		if v {
			return "1"
		} else {
			return "0"
		}
	case []byte:
		return string(v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05.999")
	default:
		return fmt.Sprint(v)
	}
}

func newColumnHeadersValueRange(columns []string) *sheets.ValueRange {
	valueRange := &sheets.ValueRange{}
	valueRange.Values = make([][]interface{}, 1)
	for _, c := range columns {
		valueRange.Values[0] = append(valueRange.Values[0], c)
	}
	return valueRange
}

func newColumnFormatRequest(sheetId int64, columns []ColumnFormatConfig) *sheets.BatchUpdateSpreadsheetRequest {
	if columns == nil {
		return nil
	}

	request := &sheets.BatchUpdateSpreadsheetRequest{}
	for _, c := range columns {
		switch c.ColumnDataType() {
		case "date":
			request.Requests = append(
				request.Requests,
				newDateFormatRequest(sheetId, int64(c.ColumnIndex()), c.ColumnFormat()),
			)
		case "money":
			request.Requests = append(
				request.Requests,
				newMoneyFormatRequest(sheetId, int64(c.ColumnIndex())),
			)
		default:
			panic(fmt.Sprintf("Unknown column format type [%v]", c.ColumnDataType()))
		}
	}

	return request
}

func newDateFormatRequest(sheetId int64, columnNumber int64, format string) *sheets.Request {
	req := newNumberFormatRequest(sheetId, columnNumber)
	req.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Type = "DATE"
	req.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Pattern = format
	return req
}

func newMoneyFormatRequest(sheetId int64, columnNumber int64) *sheets.Request {
	req := newNumberFormatRequest(sheetId, columnNumber)
	req.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Type = "NUMBER"
	req.RepeatCell.Cell.UserEnteredFormat.NumberFormat.Pattern = "#,##0.00;(#,##0.00)"
	return req
}

func newNumberFormatRequest(sheetId int64, columnNumber int64) *sheets.Request {
	req := &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetId,
				StartColumnIndex: columnNumber,
				EndColumnIndex:   columnNumber + 1,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					NumberFormat: &sheets.NumberFormat{},
				},
			},
			Fields: "userEnteredFormat.numberFormat",
		},
	}
	return req
}

func newCreateSheetRequest(sheetNo int, sheetName string, rows [][]interface{}, columns []string) (*sheets.BatchUpdateSpreadsheetRequest, error) {
	request := &sheets.Request{}
	if sheetNo == 0 {
		request.UpdateSheetProperties = newUpdateSheetRequest(sheetName, rows, columns)
	} else {
		request.AddSheet = newAddSheetRequest(sheetName, rows, columns)
	}
	batchRequest := &sheets.BatchUpdateSpreadsheetRequest{}
	batchRequest.Requests = append(batchRequest.Requests, request)
	return batchRequest, nil
}

func newAddSheetRequest(sheetName string, rows [][]interface{}, columns []string) *sheets.AddSheetRequest {
	req := &sheets.AddSheetRequest{
		Properties: newSheetProperties(sheetName, rows, columns),
	}
	return req
}

func newUpdateSheetRequest(sheetName string, rows [][]interface{}, columns []string) *sheets.UpdateSheetPropertiesRequest {
	req := &sheets.UpdateSheetPropertiesRequest{
		Fields:     "Title,GridProperties.RowCount,GridProperties.ColumnCount,GridProperties.FrozenRowCount",
		Properties: newSheetProperties(sheetName, rows, columns),
	}
	return req
}

func newSpreadSheet(documentName string) *sheets.Spreadsheet {
	s := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: documentName,
		},
	}
	return s
}

func newSheetProperties(sheetName string, rows [][]interface{}, columns []string) *sheets.SheetProperties {
	prop := &sheets.SheetProperties{
		Title: sheetName,
		GridProperties: &sheets.GridProperties{
			RowCount:          int64(len(rows)),
			ColumnCount:       int64(len(columns)),
			FrozenColumnCount: 1,
		},
	}

	return prop
}
