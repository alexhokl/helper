package datetime

import (
	"fmt"
	"time"
)

// GetLocalDateTimeString returns date and time in local timezone
func GetLocalDateTimeString(t *time.Time) string {
	loc, _ := time.LoadLocation("Local")
	lt := t.In(loc)
	return lt.Format("2006-01-02 15:04")
}

func ValidateDate(str string) error {
	_, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return fmt.Errorf("invalid date format: %v", err)
	}
	return nil
}

