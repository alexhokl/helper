package googleapi

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func NewCalendarService(ctx context.Context, oauthConfig oauth2.Config, token oauth2.Token) (*calendar.Service, error) {
	httpClient := oauthConfig.Client(ctx, &token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}
	return srv, nil
}

func GetEvents(srv *calendar.Service, calendarName string, startTime string, endTime string, query string, limit int64, orderBy string) ([]*calendar.Event, error) {
	if calendarName == "" {
		return nil, fmt.Errorf("calendar name is required")
	}

	call := srv.Events.List(calendarName).SingleEvents(true)

	if startTime != "" {
		call = call.TimeMin(startTime)
	}

	if endTime != "" {
		call = call.TimeMax(endTime)
	}

	if limit != 0 {
		call = call.MaxResults(limit)
	}

	if orderBy != "" {
		call = call.OrderBy(orderBy)
	}

	if query != "" {
		call = call.Q(query)
	}

	events, err := call.Do()
	if err != nil {
		return nil, err
	}
	if events.NextPageToken == "" {
		return events.Items, nil
	}

	return events.Items, nil
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
		Summary: summary,
		Description: description,
	}

	eventCreated, err := srv.Events.Insert(calendarID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create event: %v", err)
	}
	return eventCreated, nil
}
