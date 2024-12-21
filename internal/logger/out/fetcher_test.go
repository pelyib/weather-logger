package out

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pelyib/weather-logger/internal/shared"
)

type MockAdapter struct {
	FetchCalled  bool
	FetchInput   shared.SearchRequest
	FetchResult  []byte
	MapperCalled bool
	MapperInput  []byte
	MapperResult []shared.MeasurementResult
}

func (m *MockAdapter) SourceId() string {
	return "test.test"
}

func (m *MockAdapter) Fetch(sr shared.SearchRequest) []byte {
	m.FetchCalled = true
	m.FetchInput = sr

	if m.FetchResult == nil {
		return []byte{}
	}

	return m.FetchResult
}

func (m *MockAdapter) MapToMeasurements(rawApiRes []byte) []shared.MeasurementResult {
	m.MapperCalled = true
	m.MapperInput = rawApiRes
	if m.MapperResult == nil {
		return shared.MakeEmptyResults()
	}

	return m.MapperResult
}

func TestGetMeasurement_callsRemoteApiClientWithTheSearchRequest(t *testing.T) {
	mockAdapter := &MockAdapter{}
	fetcher := Fetcher{weatherProviderAdapter: mockAdapter}
	expectedResults := shared.MakeEmptyResults()
	searchRequest := shared.SearchRequest{Loc: shared.Location{Name: "thisisatest"}}

	results := fetcher.GetMeasurement(searchRequest)

	if !mockAdapter.FetchCalled {
		t.Errorf("Expected Fetch to be called on the adapter, but it was not.")
	}
	if searchRequest != mockAdapter.FetchInput {
		t.Errorf("Expected Fetch to be called with searchRequest, but it was not")
	}
	if results == nil || len(results) != len(expectedResults) {
		t.Errorf("Expected results to be %v, got %v", expectedResults, results)
	}
}

func TestGetMeasurement_callsMapperWithTheRemoteApiResponseBody(t *testing.T) {
	mockAdapter := &MockAdapter{}
	fetcher := Fetcher{weatherProviderAdapter: mockAdapter}
	searchRequest := shared.SearchRequest{}
	expectedResults := shared.MakeEmptyResults()
	mockAdapter.MapperResult = expectedResults
	results := fetcher.GetMeasurement(searchRequest)

	if !mockAdapter.MapperCalled {
		t.Errorf("Expected MapToMeasurements to be called on the adapter, but it was not")
	}

	if results == nil || len(results) != len(expectedResults) {
		t.Errorf("Expected results to be %v, got %v", expectedResults, results)
	}
}

func TestGetMeasurement_savesRawApiResponsesToDb(t *testing.T) {
}

func TestGetMeasurement_returnsACollectionOfMeasurements(t *testing.T) {
	mockAdapter := &MockAdapter{}
	fetcher := Fetcher{weatherProviderAdapter: mockAdapter}
	expectedResults := shared.MakeEmptyResults()
	expectedResults = append(expectedResults, shared.MeasurementResult{
		Source: "test",
		Type:   "test",
	})

	mockAdapter.MapperResult = expectedResults

	searchRequest := shared.SearchRequest{Loc: shared.Location{Name: "thisisatest"}}

	results := fetcher.GetMeasurement(searchRequest)

	if len(results) != 1 {
		t.Errorf(fmt.Sprintf("Expected the collection contains only 1 item, but it has %d items", len(results)))
	}

	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("Expected collection is different")
	}
}
