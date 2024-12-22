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
	MapperCalled bool
	MapperInput  []byte
	MapperResult []shared.MeasurementResult
}

func (m *AdapterMock) SourceId() string {
	return "test.test"
}

func (m *AdapterMock) Fetch(sr shared.SearchRequest) []byte {
	m.FetchCalled = true
	m.FetchInput = sr

	if m.FetchResult == nil {
		return []byte{}
	}

	return m.FetchResult
}

func (m *AdapterMock) MapToMeasurements(rawApiRes []byte) []shared.MeasurementResult {
	m.MapperCalled = true
	m.MapperInput = rawApiRes
	if m.MapperResult == nil {
		return shared.MakeEmptyResults()
	}

	return m.MapperResult
}

type couchDbClientMock struct {
	SaveRawApiResCalled bool
	SaveRawApiResInput  []byte
}

func (c *couchDbClientMock) saveRawApiRes(sourceId string, rawApiRes []byte) {
	c.SaveRawApiResCalled = true
	c.SaveRawApiResInput = rawApiRes
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
