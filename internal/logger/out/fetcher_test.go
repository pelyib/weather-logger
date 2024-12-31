package out

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pelyib/weather-logger/internal/shared"
)

type AdapterMock struct {
	FetchCalled  bool
	FetchInput   shared.SearchRequest
	FetchResult  []byte
	FetchError   error
	MapperCalled bool
	MapperInput  []byte
	MapperResult []shared.MeasurementResult
}

func (m *AdapterMock) sourceId() string {
	return "test.test"
}

func (m *AdapterMock) fetch(sr shared.SearchRequest) ([]byte, error) {
	m.FetchCalled = true
	m.FetchInput = sr

	if m.FetchError != nil {
		return nil, m.FetchError
	}

	if m.FetchResult == nil {
		return []byte{}, nil
	}

	return m.FetchResult, nil
}

func (m *AdapterMock) MapToMeasurements(rawApiRes []byte, loc shared.Location) ([]shared.MeasurementResult, error) {
	m.MapperCalled = true
	m.MapperInput = rawApiRes
	if m.MapperResult == nil {
		return shared.MakeEmptyResults(), nil
	}

	return m.MapperResult, nil
}

type couchDbClientMock struct {
	SaveRawApiResCalled bool
	SaveRawApiResInput  []byte
}

func (c *couchDbClientMock) saveRawApiRes(sourceId string, rawApiRes []byte) {
	c.SaveRawApiResCalled = true
	c.SaveRawApiResInput = rawApiRes
}

type loggerMock struct {
	ErrorCalled bool
	ErrorInput  string
}

func (l *loggerMock) Info(msg string)    {}
func (l *loggerMock) Warning(msg string) {}
func (l *loggerMock) Error(msg string) {
	l.ErrorCalled = true
	l.ErrorInput = msg
}

func TestGetMeasurement_callsRemoteApiClientWithTheSearchRequest(t *testing.T) {
	adapterMock := &AdapterMock{}
	couchDbClientMock := &couchDbClientMock{}
	fetcher := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
	}
	expectedResults := shared.MakeEmptyResults()
	searchRequest := shared.SearchRequest{Loc: shared.Location{Name: "thisisatest"}}

	results := fetcher.GetMeasurement(searchRequest)

	if !adapterMock.FetchCalled {
		t.Errorf("Expected Fetch to be called on the adapter, but it was not.")
	}
	if searchRequest != adapterMock.FetchInput {
		t.Errorf("Expected Fetch to be called with searchRequest, but it was not")
	}
	if results == nil || len(results) != len(expectedResults) {
		t.Errorf("Expected results to be %v, got %v", expectedResults, results)
	}
}

func TestGetMeasurement_returnsEmptyCollection_WhenFetchFails(t *testing.T) {
	adapterMock := &AdapterMock{}
	couchDbClientMock := &couchDbClientMock{}
	loggerMock := &loggerMock{}
	fetcher := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
		logger:                 loggerMock,
	}
	searchRequest := shared.SearchRequest{}
	adapterMock.FetchResult = nil
	adapterMock.FetchError = fmt.Errorf("test error")

	results := fetcher.GetMeasurement(searchRequest)
	if results == nil || len(results) != 0 {
		t.Errorf("Expected results to be empty, got %v", results)
	}

	if !loggerMock.ErrorCalled {
		t.Errorf("Expected logger to be called, but it was not")
	}
	if loggerMock.ErrorInput != "Fetching test.test failed, reason: test error" {
		t.Errorf("Expected logger to be called with 'Fetching test.test failed, reason: test error', got %s", loggerMock.ErrorInput)
	}
}

func TestGetMeasurement_callsMapperWithTheRemoteApiResponseBody(t *testing.T) {
	adapterMock := &AdapterMock{}
	couchDbClientMock := &couchDbClientMock{}
	fetcher := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
	}
	searchRequest := shared.SearchRequest{}
	expectedResults := shared.MakeEmptyResults()
	adapterMock.MapperResult = expectedResults
	results := fetcher.GetMeasurement(searchRequest)

	if !adapterMock.MapperCalled {
		t.Errorf("Expected MapToMeasurements to be called on the adapter, but it was not")
	}

	if results == nil || len(results) != len(expectedResults) {
		t.Errorf("Expected results to be %v, got %v", expectedResults, results)
	}
}

func TestGetMeasurement_savesRawApiResponsesToDb(t *testing.T) {
	adapterMock := &AdapterMock{}
	couchDbClientMock := &couchDbClientMock{}
	fetcher := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
	}

	searchRequest := shared.SearchRequest{}
	expectedResults := shared.MakeEmptyResults()
	adapterMock.MapperResult = expectedResults
	fetcher.GetMeasurement(searchRequest)

	if !couchDbClientMock.SaveRawApiResCalled {
		t.Errorf("Expected saveRawApiRes to be called on the dbClient, but it was not")
	}
}

func TestGetMeasurement_returnsACollectionOfMeasurements(t *testing.T) {
	adapterMock := &AdapterMock{}
	couchDbClientMock := &couchDbClientMock{}
	fetcher := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
	}
	expectedResults := shared.MakeEmptyResults()
	expectedResults = append(expectedResults, shared.MeasurementResult{
		Source: "test",
		Type:   "test",
	})

	adapterMock.MapperResult = expectedResults

	searchRequest := shared.SearchRequest{Loc: shared.Location{Name: "thisisatest"}}

	results := fetcher.GetMeasurement(searchRequest)

	if len(results) != 1 {
		t.Errorf(fmt.Sprintf("Expected the collection contains only 1 item, but it has %d items", len(results)))
	}

	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("Expected collection is different")
	}
}
