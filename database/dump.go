package database

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
)

// DumpTables prints the specified tables to standard output stream
func DumpTables(list []TableData) error {
	for _, data := range list {
		err := DumpTable(&data)
		if err != nil {
			return err
		}
	}
	return nil
}

// DumpTable prints the specified table to standard output stream
func DumpTable(data *TableData) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(data.Columns)
	var rows [][]string
	for _, r := range data.Rows {
		vals := getStringValues(r)
		rows = append(rows, vals)
	}
	table.AppendBulk(rows)
	return nil
}

func getStringValues(row []interface{}) []string {
	list := []string{}
	for _, c := range row {
		list = append(list, getStringValue(c.(*interface{})))
	}
	return list
}

func getStringValue(val *interface{}) string {
	switch v := (*val).(type) {
	case nil:
		return "NULL"
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case []byte:
		return string(v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05.999")
	default:
		return fmt.Sprint(v)
	}
}
