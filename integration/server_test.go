package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSomeData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Get("http://localhost:8080/api/v1/some-data?key=KanchyEnjoyers")
	}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/some-data?key=KanchyEnjoyers")
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %v", resp.Status)
	}

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if data["key"] != "KanchyEnjoyers" || data["value"] == "" {
		t.Fatalf("unexpected response: %v", data)
	}
}
