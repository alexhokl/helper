package googleapi

import (
	"testing"
	"time"
)

var testTime = time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)

func TestGetEventsEmptyCalendarID(t *testing.T) {
	// GetEvents should return error when calendarID is empty
	_, err := GetEvents(nil, "", "", "", "", 10, "")
	if err == nil {
		t.Error("GetEvents() with empty calendarID should return error")
	}
}

func TestCreateEventEmptyCalendarID(t *testing.T) {
	_, err := CreateEvent(nil, "", "summary", "description", "2023-01-01T00:00:00Z", "2023-01-01T01:00:00Z")
	if err == nil {
		t.Error("CreateEvent() with empty calendarID should return error")
	}
}

func TestDeleteEventEmptyCalendarID(t *testing.T) {
	err := DeleteEvent(nil, "", "event-id")
	if err == nil {
		t.Error("DeleteEvent() with empty calendarID should return error")
	}
}

func TestDeleteEventEmptyEventID(t *testing.T) {
	err := DeleteEvent(nil, "calendar-id", "")
	if err == nil {
		t.Error("DeleteEvent() with empty eventID should return error")
	}
}

func TestDeleteEventBothEmpty(t *testing.T) {
	err := DeleteEvent(nil, "", "")
	if err == nil {
		t.Error("DeleteEvent() with empty calendarID and eventID should return error")
	}
}

func TestDeleteEventsEmptySlice(t *testing.T) {
	// DeleteEvents with empty slice should succeed (no-op)
	err := DeleteEvents(nil, "calendar-id", []string{})
	if err != nil {
		t.Errorf("DeleteEvents() with empty slice error: %v", err)
	}
}

func TestGetCalendarTimeZoneEmptyCalendarID(t *testing.T) {
	_, err := GetCalendarTimeZone(nil, "")
	if err == nil {
		t.Error("GetCalendarTimeZone() with empty calendarID should return error")
	}
}

func TestPatchEventDatesEmptyCalendarID(t *testing.T) {
	_, err := PatchEventDates(nil, "", "event-id", testTime, testTime)
	if err == nil {
		t.Error("PatchEventDates() with empty calendarID should return error")
	}
}

func TestPatchEventDatesEmptyEventID(t *testing.T) {
	_, err := PatchEventDates(nil, "calendar-id", "", testTime, testTime)
	if err == nil {
		t.Error("PatchEventDates() with empty eventID should return error")
	}
}

func TestListDefaultMax(t *testing.T) {
	if listDefaultMax != 250 {
		t.Errorf("listDefaultMax = %d, want 250", listDefaultMax)
	}
}
