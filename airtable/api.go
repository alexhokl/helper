package airtable

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alexhokl/helper/httphelper"
	"github.com/alexhokl/helper/jsonhelper"
)

const defaultMaxRecords = 100
const contentTypeJSON = "application/json"
const apiBasePath = "https://api.airtable.com/v0"
const defaultOffset = "first_call"

// UpdateRecordsRequest returns a request to update records of Airtable
func UpdateRecordsRequest(baseID string, tableName string, patchBody *bytes.Buffer, ctx context.Context) (*http.Request, error) {
	path := fmt.Sprintf("%s/%s/%s", apiBasePath, baseID, tableName)
	headers := map[string]string{
		"Content-Type": contentTypeJSON,
	}

	request, err := httphelper.NewRequest(http.MethodPatch, path, nil, headers, patchBody)
	if err != nil {
		return nil, err
	}

	if ctx != nil {
		request = request.WithContext(ctx)
	}

	return request, nil
}

// ListRecords returns a list of records from Airtable
func ListRecords[T AirtableFields](httpClient *http.Client, baseID string, tableName string, viewName string, ctx context.Context, maxRecords int) ([]*AirtableRecord[T], error) {
	var items []*AirtableRecord[T]

	offset := defaultOffset

	for offset != "" {
		request, err := listRecordsRequest(baseID, tableName, viewName, ctx, maxRecords, offset)
		if err != nil {
			return nil, fmt.Errorf("unable to create request: %v", err)
		}
		response, err := httpClient.Do(request)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return nil, err
		}

		if !httphelper.IsSuccessResponse(response) {
			return nil, handleErrorResponse(response.StatusCode, body)
		}

		if !httphelper.HasContentType(response, contentTypeJSON) {
			return nil, fmt.Errorf("Content-Type is not %s", contentTypeJSON)
		}

		var list AirtableRecords[T]
		if err := jsonhelper.ParseJSONFromBytes(&list, body); err != nil {
			return nil, err
		}

		for i := range list.Records {
			items = append(items, &list.Records[i])
		}
		offset = list.Offset
	}
	return items, nil
}

// UpdateRecords updates records of Airtable and returns the records updated
func UpdateRecords[T AirtableFields](httpClient *http.Client, request *http.Request) ([]*AirtableRecord[T], error) {
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	if !httphelper.IsSuccessResponse(response) {
		return nil, handleErrorResponse(response.StatusCode, body)
	}

	if !httphelper.HasContentType(response, contentTypeJSON) {
		return nil, fmt.Errorf("Content-Type is not %s", contentTypeJSON)
	}

	var list AirtableRecords[T]
	if err := jsonhelper.ParseJSONFromBytes(&list, body); err != nil {
		return nil, err
	}

	var items []*AirtableRecord[T]
	for i := range list.Records {
		items = append(items, &list.Records[i])
	}
	return items, nil
}

func CreateRecord[Tin AirtableFields, T AirtableFields](httpClient *http.Client, record *Tin, baseID string, tableName string, ctx context.Context) ([]*AirtableRecord[T], error) {
	path := fmt.Sprintf("%s/%s/%s", apiBasePath, baseID, tableName)
	headers := map[string]string{
		"Content-Type": contentTypeJSON,
	}

	viewModel := CreateRecordsRequest[Tin]{}
	viewModel.Records = append(
		viewModel.Records,
		CreateRecordRequest[Tin]{
			Fields: *record,
		},
	)
	viewModel.Typecast = true
	body, err := EncodePostAsJSON(viewModel)
	if err != nil {
		return nil, err
	}

	request, err := httphelper.NewRequest(http.MethodPost, path, nil, headers, body)
	if err != nil {
		return nil, err
	}

	if ctx != nil {
		request = request.WithContext(ctx)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	responseBody, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	if !httphelper.IsSuccessResponse(response) {
		return nil, handleErrorResponse(response.StatusCode, responseBody)
	}

	if !httphelper.HasContentType(response, contentTypeJSON) {
		return nil, fmt.Errorf("Content-Type is not %s", contentTypeJSON)
	}

	var list AirtableRecords[T]
	if err := jsonhelper.ParseJSONFromBytes(&list, responseBody); err != nil {
		return nil, err
	}

	var items []*AirtableRecord[T]
	for i := range list.Records {
		items = append(items, &list.Records[i])
	}
	return items, nil
}

// EncodePostAsJSON returns bytes of JSON encoded patch request
func EncodePostAsJSON[T AirtableFields](req CreateRecordsRequest[T]) (bodyBuf *bytes.Buffer, err error) {
	bodyBuf = &bytes.Buffer{}
	err = json.NewEncoder(bodyBuf).Encode(req)
	if err != nil {
		return nil, err
	}

	return bodyBuf, nil
}

// EncodePatchAsJSON returns bytes of JSON encoded patch request
func EncodePatchAsJSON[T any](patchRequest PatchItemsRequest[T]) (bodyBuf *bytes.Buffer, err error) {
	bodyBuf = &bytes.Buffer{}
	err = json.NewEncoder(bodyBuf).Encode(patchRequest)
	if err != nil {
		return nil, err
	}

	return bodyBuf, nil
}

func handleErrorResponse(statusCode int, responseBody []byte) error {
	switch statusCode {
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limited by airtable API")
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized to access airtable API; possible malformed access token")
	case http.StatusForbidden:
		return getError(responseBody, "forbidden to access airtable API")
	case http.StatusNotFound:
		return getError(responseBody, "not found")
	case http.StatusRequestEntityTooLarge:
		return fmt.Errorf("request body too large")
	case http.StatusUnprocessableEntity:
		return getError(responseBody, "unprocessable entity")
	default:
		return fmt.Errorf("API error: %d [%s]", statusCode, string(responseBody))
	}
}

func getError(responseBody []byte, message string) error {
	var errorResponse ErrorResponse
	if err := jsonhelper.ParseJSONFromBytes(&errorResponse, responseBody); err != nil {
		var simpleErrorResponse SimpleErrorResponse
		if err := jsonhelper.ParseJSONFromBytes(&simpleErrorResponse, responseBody); err != nil {
			return err
		}
		return fmt.Errorf("%s: %s", message, simpleErrorResponse.Error)
	}
	return getErrorFromErrorResponse(errorResponse, message)
}

func getErrorFromErrorResponse(errorResponse ErrorResponse, message string) error {
	if errorResponse.Error.Message != "" {
		return fmt.Errorf("%s: error type [%s], message [%s]", message, errorResponse.Error.Type, errorResponse.Error.Message)
	}
	return fmt.Errorf("%s: error type [%s]", message, errorResponse.Error.Type)
}

func listRecordsRequest(baseID string, tableName string, viewName string, ctx context.Context, maxRecords int, offset string) (*http.Request, error) {
	path := fmt.Sprintf("%s/%s/%s", apiBasePath, baseID, tableName)
	headers := map[string]string{
		"Content-Type": contentTypeJSON,
	}

	if maxRecords == 0 {
		maxRecords = defaultMaxRecords
	}

	queryParams := map[string]string{
		"maxRecords": fmt.Sprintf("%d", maxRecords),
		"view":       viewName,
	}
	if offset != defaultOffset && offset != "" {
		queryParams["offset"] = offset
	}

	request, err := httphelper.NewRequest(http.MethodGet, path, queryParams, headers, nil)
	if err != nil {
		return nil, err
	}

	if ctx != nil {
		request = request.WithContext(ctx)
	}

	return request, nil
}
