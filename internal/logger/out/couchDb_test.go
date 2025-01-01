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
)

func TestSaveRawApiRes_returnsError_whenNoDbConfigGiven(t *testing.T) {
	c := client{
		now: func() time.Time {
			now, _ := time.Parse("2006-01-02 03:04:05", "2024-12-12 10:10:10")
			return now
		},
		config: Config{
			Host: "http://example.com",
			Dbs: map[string]Db{
				"banan": {
					Name:     "api_raw_responses",
					User:     "logger",
					Password: "logger",
				},
			},
		}}
	rawApiRes := []byte("{\"key\": \"value\"}")

	err := c.saveRawApiRes("source-id", rawApiRes)

	if err == nil {
		t.Fatal("Expected error, but got nothing")
	}
}

func TestSaveRawApiRes_returnsError_whenRawApiResIsIvalidJson(t *testing.T) {
	c := client{
		now: func() time.Time {
			now, _ := time.Parse("2006-01-02 03:04:05", "2024-12-12 10:10:10")
			return now
		},
		config: Config{
			Host: "https://not-existing-domain.com",
			Dbs: map[string]Db{
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

func TestSaveRawApiRes_returnsError_whenDbNotReachable(t *testing.T) {
	c := client{
		now: func() time.Time {
			now, _ := time.Parse("2006-01-02 03:04:05", "2024-12-12 10:10:10")
			return now
		},
		config: Config{
			Host: "https://not-existing-domain.com",
			Dbs: map[string]Db{
				"api_raw_responses": {
					Name:     "api_raw_responses",
					User:     "logger",
					Password: "logger",
				},
			},
		}}

	rawApiRes := []byte("{\"key\": \"value\"}")

	err := c.saveRawApiRes("source-id", rawApiRes)

	if err == nil {
		t.Fatal("Expected error but got nothing")
	}
}

func TestSaveRawApiRes_returnsError_whenDbCallIsUnsuccesful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	c := client{
		now: func() time.Time {
			now, _ := time.Parse("2006-01-02 03:04:05", "2024-12-12 10:10:10")
			return now
		},
		config: Config{
			Host: ts.URL,
			Dbs: map[string]Db{
				"api_raw_responses": {
					Name:     "api_raw_responses",
					User:     "logger",
					Password: "logger",
				},
			},
		}}

	rawApiRes := []byte("{\"key\": \"value\"}")

	err := c.saveRawApiRes("source-id", rawApiRes)

	if err == nil {
		t.Fatal("Expected error but got nothing")
	}
}

func TestSaveRawApiRes_callsDbEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api_raw_responses/") {
			t.Fatalf("Expected path mismatch, got %s", r.URL.Path)
		}

		bodyAsBytes, _ := io.ReadAll(r.Body)
		bodyAsString := string(bodyAsBytes)

		expectedBody := "{\"obj\":{\"calledAt\":\"2024-12-12T10:10:10Z\",\"sourceId\":\"source-id\",\"raw\":{\"key\":\"value\"}}}"
		if expectedBody != bodyAsString {
			t.Fatalf("Expected body mismatch, got %s", bodyAsString)
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Fatal("Expected auth header is set, but it is not")
		}
		if !strings.HasPrefix(authHeader, "Basic") {
			t.Fatalf("Expected auth header is Basic, got %s", authHeader)
		}

		w.WriteHeader(200)
		w.Write([]byte(`desired response here`))
	}))
	defer ts.Close()
	c := client{
		now: func() time.Time {
			now, _ := time.Parse("2006-01-02 03:04:05", "2024-12-12 10:10:10")
			return now
		},
		config: Config{
			Host: ts.URL,
			Dbs: map[string]Db{
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
