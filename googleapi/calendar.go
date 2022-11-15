package googleapi

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const listDefaultMax = 250
const listNoLimit = 0

func NewCalendarService(ctx context.Context, oauthConfig oauth2.Config, token oauth2.Token) (*calendar.Service, error) {
	httpClient := oauthConfig.Client(ctx, &token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}
	return srv, nil
}

func GetEvents(srv *calendar.Service, calendarID string, startTime string, endTime string, query string, limit int, orderBy string) ([]*calendar.Event, error) {
	if calendarID == "" {
		return nil, fmt.Errorf("calendar ID is required")
	}

	call := srv.Events.List(calendarID).SingleEvents(true)

	if startTime != "" {
		call = call.TimeMin(startTime)
	}

	if endTime != "" {
		call = call.TimeMax(endTime)
	}

	if orderBy != "" {
		call = call.OrderBy(orderBy)
	}

	if query != "" {
		call = call.Q(query)
	}

	events, err := listEvents(call, limit)
	if err != nil {
		return nil, err
	}

	count := len(events.Items)

	var allEvents []*calendar.Event
	allEvents = append(allEvents, events.Items...)

	for events.NextPageToken != "" && count < limit {
		nextPageToken := events.NextPageToken
		call = call.PageToken(nextPageToken)
		remainingLimit := limit - count
		events, err = listEvents(call, remainingLimit)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve results of next page [%s]: %v", nextPageToken, err)
		}
		allEvents = append(allEvents, events.Items...)
	}

	return allEvents, nil
}

func GetCalendars(srv *calendar.Service) ([]*calendar.CalendarListEntry, error) {
	list, err := srv.CalendarList.List().Do()
	if err != nil {
		return nil, err
	}
	if list.NextPageToken == "" {
		return list.Items, nil
	}
	return list.Items, nil
}

func CreateEvent(srv *calendar.Service, calendarID string, summary string, description string, startDateTime string, endDateTime string) (*calendar.Event, error) {
	if calendarID == "" {
		return nil, fmt.Errorf("calendar name is required")
	}

	event := &calendar.Event{
		Start: &calendar.EventDateTime{
			DateTime: startDateTime,
		},
		End: &calendar.EventDateTime{
			DateTime: endDateTime,
		},
		Summary:     summary,
		Description: description,
	}

	eventCreated, err := srv.Events.Insert(calendarID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create event: %v", err)
	}
	return eventCreated, nil
}

func DeleteEvent(srv *calendar.Service, calendarID string, eventID string) error {
	if calendarID == "" {
		return fmt.Errorf("calendar name is required")
	}

	if eventID == "" {
		return fmt.Errorf("event id is required")
	}

	err := srv.Events.Delete(calendarID, eventID).Do()
	if err != nil {
		return fmt.Errorf("unable to delete event: %v", err)
	}
	return nil
}

func DeleteEvents(srv *calendar.Service, calendarID string, eventIDs []string) error {
	for _, eventID := range eventIDs {
		if err := DeleteEvent(srv, calendarID, eventID); err != nil {
			return err
		}
	}
	return nil
}

// / GetCalendarTimeZone
func GetCalendarTimeZone(srv *calendar.Service, calendarID string) (*time.Location, error) {
	if calendarID == "" {
		return nil, fmt.Errorf("calendar name is required")
	}

	calendar, err := srv.Calendars.Get(calendarID).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get calendar: %v", err)
	}
	ianaTimeZone := calendar.TimeZone
	return time.LoadLocation(ianaTimeZone)
}

func PatchEventDates(srv *calendar.Service, calendarID string, eventID string, start time.Time, end time.Time) (*calendar.Event, error) {
	if calendarID == "" {
		return nil, fmt.Errorf("calendar name is required")
	}

	if eventID == "" {
		return nil, fmt.Errorf("event id is required")
	}

	event := &calendar.Event{
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
		},
	}

	return srv.Events.Patch(calendarID, eventID, event).Do()
}

func listEvents(call *calendar.EventsListCall, limit int) (*calendar.Events, error) {
	if limit < listDefaultMax {
		call = call.MaxResults(int64(limit))
	}
	return call.Do()
}
