package out

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

func Test_put_returnsError_whenNoDbConfigGiven(t *testing.T) {
	c := client{
		now: func() time.Time {
			return time.Date(2024, 12, 12, 10, 10, 10, 0, time.UTC)
		},
		config: shared.CouchDb{
			Host: "http://example.com",
			Dbs: map[string]shared.Db{
				"banan": {
					Name:     "api_raw_responses",
					User:     "logger",
					Password: "logger",
				},
			},
		}}
	rawApiRes := []byte("{\"key\": \"value\"}")

	err := c.put("api_raw_responses", rawApiRes)

	if err == nil {
		t.Fatal("Expected error, but got nothing")
	}

	if err.Error() != "No DB config specified for api_raw_responses" {
		t.Fatalf("Expected error message mismatch, got %s", err.Error())
	}
}

func Test_put_returnsError_whenDbNotReachable(t *testing.T) {
	c := client{
		now: func() time.Time {
			return time.Date(2024, 12, 12, 10, 10, 10, 0, time.UTC)
		},
		config: shared.CouchDb{
			Host: "https://not-existing-domain.com",
			Dbs: map[string]shared.Db{
				"api_raw_responses": {
					Name:     "api_raw_responses",
					User:     "logger",
					Password: "logger",
				},
			},
		}}

	rawApiRes := []byte("{\"key\": \"value\"}")

	err := c.put("api_raw_responses", rawApiRes)

	if err == nil {
		t.Fatal("Expected error but got nothing")
	}

	if !strings.Contains(err.Error(), "dial tcp: lookup not-existing-domain.com: no such host") {
		t.Fatalf("Expected error message mismatch, got %s", err.Error())
	}
}

func Test_put_returnsError_whenDbCallIsUnsuccesful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	c := client{
		now: func() time.Time {
			return time.Date(2024, 12, 12, 10, 10, 10, 0, time.UTC)
		},
		config: shared.CouchDb{
			Host: ts.URL,
			Dbs: map[string]shared.Db{
				"api_raw_responses": {
					Name:     "api_raw_responses",
					User:     "logger",
					Password: "logger",
				},
			},
		}}

	rawApiRes := []byte("{\"key\": \"value\"}")

	err := c.put("api_raw_responses", rawApiRes)

	if err == nil {
		t.Fatal("Expected error but got nothing")
	}
}

func Test_put_callsDbEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api_raw_responses/") {
			t.Fatalf("Expected path mismatch, got %s", r.URL.Path)
		}

		bodyAsBytes, _ := io.ReadAll(r.Body)
		bodyAsString := string(bodyAsBytes)

		expectedBody := "{\"key\":\"value\"}"
		if expectedBody != bodyAsString {
			t.Fatalf("Expected body mismatch, got %s", bodyAsString)
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "Basic bG9nZ2VyOmxvZ2dlcg==" {
			t.Fatalf("Expected auth header mismatch, got %s", authHeader)
		}

		w.WriteHeader(200)
		w.Write([]byte(`desired response here`))
	}))
	defer ts.Close()
	c := client{
		now: func() time.Time {
			return time.Date(2024, 12, 12, 10, 10, 10, 0, time.UTC)
		},
		config: shared.CouchDb{
			Host: ts.URL,
			Dbs: map[string]shared.Db{
				"api_raw_responses": {
					Name:     "api_raw_responses",
					User:     "logger",
					Password: "logger",
				},
			},
		}}

	rawApiRes := []byte("{\"key\":\"value\"}")

	err := c.put("api_raw_responses", rawApiRes)

	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
}

func Test_saveRawApiRes_returnsError_whenRawApiResIsInvalidJson(t *testing.T) {
	c := client{
		now: func() time.Time {
			now, _ := time.Parse("2006-01-02 03:04:05", "2024-12-12 10:10:10")
			return now
		},
		config: shared.CouchDb{
			Host: "https://not-important-domain.com",
			Dbs: map[string]shared.Db{
				"api_raw_responses": {
					Name:     "api_raw_responses",
					User:     "logger",
					Password: "logger",
				},
			},
		}}

	invalidJSON := []byte("{invalid-json}")

	err := c.saveRawApiRes("source-id", invalidJSON)

	if err == nil {
		t.Fatal("Expected error, but got nil")
	}

	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Fatalf("Expected json.SyntaxError, but got %v", err)
	}
}

func Test_saveRawApiRes_callsPutWithCorrectData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api_raw_responses/") {
			t.Fatalf("Expected path mismatch, got %s", r.URL.Path)

			w.WriteHeader(404)
			w.Write([]byte(`not found`))
			return
		}

		bodyAsBytes, _ := io.ReadAll(r.Body)
		bodyAsString := string(bodyAsBytes)

		expectedBody := "{\"obj\":{\"calledAt\":\"2024-12-12T10:10:10Z\",\"sourceId\":\"source-id\",\"raw\":{\"key\":\"value\"}}}"
		if expectedBody != bodyAsString {
			t.Fatalf("Expected body mismatch, got %s", bodyAsString)

			w.WriteHeader(400)
			return
		}

		w.WriteHeader(200)
		w.Write([]byte(`desired response here`))
	}))
	defer ts.Close()
	c := client{
		now: func() time.Time {
			return time.Date(2024, 12, 12, 10, 10, 10, 0, time.UTC)
		},
		config: shared.CouchDb{
			Host: ts.URL,
			Dbs: map[string]shared.Db{
				"api_raw_responses": {
					Name:     "api_raw_responses",
					User:     "logger",
					Password: "logger",
				},
			},
		}}

	rawApiRes := []byte("{\"key\": \"value\"}")

	err := c.saveRawApiRes("source-id", rawApiRes)

	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
}

func Test_saveMeasurement_callsPutForEveryMeasurement(t *testing.T) {
	called := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/metrics/") {
			t.Fatalf("Expected path mismatch, got %s", r.URL.Path)

			w.WriteHeader(404)
			w.Write([]byte(`not found`))
			return
		}

		bodyAsBytes, _ := io.ReadAll(r.Body)
		bodyAsString := string(bodyAsBytes)

		expectedBodies := []string{
			"\"obj\":{\"source\":\"test\",\"type\":\"test\",\"min\":1,\"max\":2,",
			"\"obj\":{\"source\":\"test\",\"type\":\"test\",\"min\":3,\"max\":4,",
		}
		bodyFound := false

		for _, expectedBody := range expectedBodies {
			if strings.Contains(bodyAsString, expectedBody) {
				bodyFound = true
			}
		}

		if !bodyFound {
			t.Fatalf("Expected body mismatch, got %s", bodyAsString)

			w.WriteHeader(400)
			return
		}

		called++
		w.WriteHeader(200)
		w.Write([]byte(`desired response here`))
	}))
	defer ts.Close()
	c := client{
		now: func() time.Time {
			return time.Date(2024, 12, 12, 10, 10, 10, 0, time.UTC)
		},
		config: shared.CouchDb{
			Host: ts.URL,
			Dbs: map[string]shared.Db{
				"metrics": {
					Name:     "metrics",
					User:     "logger",
					Password: "logger",
				},
			},
		}}

	measurements := []shared.MeasurementResult{
		{Source: "test", Type: "test", Min: 1, Max: 2},
		{Source: "test", Type: "test", Min: 3, Max: 4},
	}

	err := c.saveMeasurements(measurements)

	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if called != 2 {
		t.Fatalf("Expected 2 calls, got %d", called)
	}
}

func Test_saveMeasurement_returnsError_ifAnyCallFails(t *testing.T) {
	called := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/metrics/") {
			t.Fatalf("Expected path mismatch, got %s", r.URL.Path)

			w.WriteHeader(404)
			w.Write([]byte(`not found`))
			return
		}

		bodyAsBytes, _ := io.ReadAll(r.Body)
		bodyAsString := string(bodyAsBytes)

		expectedBodies := []string{
			"\"obj\":{\"source\":\"test\",\"type\":\"test\",\"min\":1,\"max\":2,",
			"\"obj\":{\"source\":\"test\",\"type\":\"test\",\"min\":3,\"max\":4,",
		}
		bodyFound := false

		for _, expectedBody := range expectedBodies {
			if strings.Contains(bodyAsString, expectedBody) {
				bodyFound = true
			}
		}

		if !bodyFound {
			t.Fatalf("Expected body mismatch, got %s", bodyAsString)

			w.WriteHeader(400)
			return
		}

		called++

		if called == 2 {
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)
		w.Write([]byte(`desired response here`))
	}))
	defer ts.Close()
	c := client{
		now: func() time.Time {
			return time.Date(2024, 12, 12, 10, 10, 10, 0, time.UTC)
		},
		config: shared.CouchDb{
			Host: ts.URL,
			Dbs: map[string]shared.Db{
				"metrics": {
					Name:     "metrics",
					User:     "logger",
					Password: "logger",
				},
			},
		}}

	measurements := []shared.MeasurementResult{
		{Source: "test", Type: "test", Min: 1, Max: 2},
		{Source: "test", Type: "test", Min: 3, Max: 4},
	}

	err := c.saveMeasurements(measurements)

	if err == nil {
		t.Fatal("Expected error, but got nothing")
	}

	if err.Error() != "Could not save, status code: 500" {
		t.Fatalf("Expected error message mismatch, got %s", err.Error())
	}

	if called != 2 {
		t.Fatalf("Expected 2 calls, got %d", called)
	}
}
