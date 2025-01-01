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
	MapperError  error
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

func (m *AdapterMock) mapToMeasurements(rawApiRes []byte, loc shared.Location) ([]shared.MeasurementResult, error) {
	m.MapperCalled = true
	m.MapperInput = rawApiRes
	if m.MapperResult == nil && m.MapperError == nil {
		return shared.MakeEmptyResults(), nil
	}

	if m.MapperError != nil {
		return nil, m.MapperError
	}

	return m.MapperResult, nil
}

type couchDbClientMock struct {
	SaveRawApiResCalled bool
	SaveRawApiResInput  []byte
	SaveRawApiResResult error
}

func (c *couchDbClientMock) saveRawApiRes(sourceId string, rawApiRes []byte) error {
	c.SaveRawApiResCalled = true
	c.SaveRawApiResInput = rawApiRes

	if c.SaveRawApiResResult != nil {
		return c.SaveRawApiResResult
	}

	return nil
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
	sut := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
	}
	expectedResults := shared.MakeEmptyResults()
	searchRequest := shared.SearchRequest{Loc: shared.Location{Name: "thisisatest"}}

	actual := sut.GetMeasurement(searchRequest)

	if !adapterMock.FetchCalled {
		t.Errorf("Expected Fetch to be called on the adapter, but it was not.")
	}
	if searchRequest != adapterMock.FetchInput {
		t.Errorf("Expected Fetch to be called with searchRequest, but it was not")
	}
	if actual == nil || len(actual) != len(expectedResults) {
		t.Errorf("Expected results to be %v, got %v", expectedResults, actual)
	}
}

func TestGetMeasurement_returnsEmptyCollection_WhenFetchFails(t *testing.T) {
	adapterMock := &AdapterMock{
		FetchResult: nil,
		FetchError:  fmt.Errorf("test error"),
	}
	couchDbClientMock := &couchDbClientMock{}
	loggerMock := &loggerMock{}
	sut := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
		logger:                 loggerMock,
	}

	actual := sut.GetMeasurement(shared.SearchRequest{})

	if actual == nil || len(actual) != 0 {
		t.Errorf("Expected results to be empty, got %v", actual)
	}

	if !loggerMock.ErrorCalled {
		t.Errorf("Expected logger to be called, but it was not")
	}
	if loggerMock.ErrorInput != "Fetching test.test failed, reason: test error" {
		t.Errorf("Expected logger to be called with 'Fetching test.test failed, reason: test error', got %s", loggerMock.ErrorInput)
	}
}

func TestGetMeasurement_callsMapperWithTheRemoteApiResponseBody(t *testing.T) {
	expectedResults := shared.MakeEmptyResults()
	adapterMock := &AdapterMock{MapperResult: expectedResults}
	couchDbClientMock := &couchDbClientMock{}

	sut := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
	}

	actual := sut.GetMeasurement(shared.SearchRequest{})

	if !adapterMock.MapperCalled {
		t.Errorf("Expected MapToMeasurements to be called on the adapter, but it was not")
	}

	if actual == nil || len(actual) != len(expectedResults) {
		t.Errorf("Expected results to be %v, got %v", expectedResults, actual)
	}
}

func TestGetMeasurement_logsError_whenCouldNotSaveRawApiResponse(t *testing.T) {
	adapterMock := &AdapterMock{}
	couchDbClientMock := &couchDbClientMock{
		SaveRawApiResResult: fmt.Errorf("test error"),
	}
	loggerMock := &loggerMock{}

	sut := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
		logger:                 loggerMock,
	}

	searchRequest := shared.SearchRequest{}
	expectedResults := shared.MakeEmptyResults()
	adapterMock.MapperResult = expectedResults

	sut.GetMeasurement(searchRequest)

	if !couchDbClientMock.SaveRawApiResCalled {
		t.Errorf("Expected saveRawApiRes to be called on the dbClient, but it was not")
	}

	if !loggerMock.ErrorCalled {
		t.Errorf("Expected logger to be called, but it was not")
	}

	if loggerMock.ErrorInput != "Saving raw API response failed, reason: test error" {
		t.Errorf("Expected logger message mismatch, got %s", loggerMock.ErrorInput)
	}
}

func TestGetMeasurement_savesRawApiResponsesToDb(t *testing.T) {
	adapterMock := &AdapterMock{
		MapperResult: shared.MakeEmptyResults(),
	}
	couchDbClientMock := &couchDbClientMock{}

	sut := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
	}

	sut.GetMeasurement(shared.SearchRequest{})

	if !couchDbClientMock.SaveRawApiResCalled {
		t.Errorf("Expected saveRawApiRes to be called on the dbClient, but it was not")
	}
}

func TestGetMeasurement_returnsEmptyCollection_whenMapperReturnsNilInCaseOfError(t *testing.T) {
	adapterMock := &AdapterMock{
		MapperResult: nil,
		MapperError:  fmt.Errorf("test error"),
	}
	couchDbClientMock := &couchDbClientMock{}
	loggerMock := &loggerMock{}

	sut := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
		logger:                 loggerMock,
	}

	actual := sut.GetMeasurement(shared.SearchRequest{})

	if len(actual) != 0 {
		t.Errorf("Expected results to be empty, got %v", actual)
	}

	if !loggerMock.ErrorCalled {
		t.Errorf("Expected logger to be called, but it was not")
	}

	if loggerMock.ErrorInput != "Mapping raw API response to measurements failed, reason: test error" {
		t.Errorf("Expected logger message mismatch, got %s", loggerMock.ErrorInput)
	}
}

func TestGetMeasurement_returnsACollectionOfMeasurements(t *testing.T) {
	expectedResults := shared.MakeEmptyResults()
	expectedResults = append(expectedResults, shared.MeasurementResult{
		Source: "test",
		Type:   "test",
	})
	adapterMock := &AdapterMock{
		MapperResult: expectedResults,
	}

	couchDbClientMock := &couchDbClientMock{}

	sut := Fetcher{
		weatherProviderAdapter: adapterMock,
		dbClient:               couchDbClientMock,
	}

	searchRequest := shared.SearchRequest{Loc: shared.Location{Name: "thisisatest"}}

	actual := sut.GetMeasurement(searchRequest)

	if len(actual) != 1 {
		t.Errorf(fmt.Sprintf("Expected the collection contains only 1 item, but it has %d items", len(actual)))
	}

	if !reflect.DeepEqual(actual, expectedResults) {
		t.Errorf("Expected collection is different")
	}
}
