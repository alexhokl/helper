package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

// TableData struct
type TableData struct {
	Rows    [][]interface{}
	Columns []string
}

// GetConnection returns a SQL database connection
func GetConnection(config *Config) (*sql.DB, error) {
	parameters := url.Values{}
	parameters.Add("database", config.Name)
	connectionURL := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(config.Username, config.Password),
		Host:     fmt.Sprintf("%s:%d", config.Server, config.Port),
		RawQuery: parameters.Encode(),
	}
	conn, err := sql.Open("sqlserver", connectionURL.String())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetData returns data retrieved by using query with conn
func GetData(conn *sql.DB, query string) (*TableData, error) {
	rows, errQuery := conn.Query(query)
	if errQuery != nil {
		return nil, errQuery
	}
	defer rows.Close()

	cols, errColumns := rows.Columns()
	if errColumns != nil {
		return nil, errColumns
	}

	columnCount := len(cols)

	var dataRows [][]interface{}
	for rows.Next() {
		vals := make([]interface{}, columnCount)
		for i := 0; i < columnCount; i++ {
			vals[i] = new(interface{})
		}
		err := rows.Scan(vals...)
		if err != nil {
			return nil, err
		}
		dataRows = append(dataRows, vals)
	}

	data := &TableData{
		Rows:    dataRows,
		Columns: cols,
	}

	return data, nil
}

// GetValue returns a typed value from a cell reference
func GetValue(pval *interface{}) string {
	switch v := (*pval).(type) {
	case nil:
		return "NULL"
	case bool:
		if v {
			return "1"
		}
		return "0"
	case []byte:
		return string(v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05.999")
	default:
		return fmt.Sprint(v)
	}
}
