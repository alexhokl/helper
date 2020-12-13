package googleapi

import (
	"context"
	"errors"
	"fmt"

	"googlemaps.github.io/maps"
)

func GetCoordinatesFromStreetAddress(client *maps.Client, address string) (longitude float64, latitude float64, throwErr error) {
	req := &maps.GeocodingRequest{
		Address: address,
	}
	results, err := client.Geocode(context.Background(), req)
	if err != nil {
		throwErr = err
		return
	}
	if len(results) == 0 {
		throwErr = errors.New(fmt.Sprintf("Unable to resolve address [%s]", address))
		return
	}
	longitude = results[0].Geometry.Location.Lng
	latitude = results[0].Geometry.Location.Lat
	return
}
