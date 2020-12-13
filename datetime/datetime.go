package datetime

import "time"

// GetLocalDateTimeString returns date and time in local timezone
func GetLocalDateTimeString(t *time.Time) string {
	loc, _ := time.LoadLocation("Local")
	lt := t.In(loc)
	return lt.Format("2006-01-02 15:04")
}
