package database

import (
	"database/sql"
	"fmt"

	"github.com/alexhokl/helper/iohelper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetDatabaseDailector(pathDatabaseConnectionString string) (gorm.Dialector, error) {
	if pathDatabaseConnectionString == "" {
		return nil, fmt.Errorf("file path to database connection string is not set")
	}
	connectionString, err := iohelper.ReadFirstLineFromFile(pathDatabaseConnectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to read password: %w", err)
	}
	if connectionString == "" {
		return nil, fmt.Errorf("database connection string is empty")
	}

	return postgres.Open(connectionString), nil
}

func GetDatabaseDialectorFromConnection(conn *sql.DB) gorm.Dialector {
	return postgres.New(postgres.Config{
		Conn: conn,
		DriverName: "postgres",
	})
}

func GetDatabaseConnection(dialector gorm.Dialector) (*gorm.DB, error) {
	conn, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return conn, nil
}
