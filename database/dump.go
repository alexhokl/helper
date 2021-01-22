package database

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
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
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "%s\t\n", strings.Join(data.Columns, "\t"))
	for _, r := range data.Rows {
		vals := getStringValues(r)
		fmt.Fprintf(w, "%s\t\n", strings.Join(vals, "\t"))
	}
	w.Flush()
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
