package main

import (
	"context"
	"hash/fnv"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealth_HealthyServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	assert.True(t, health(server.Listener.Addr().String()), "Expected health check to pass for a healthy server")
}

func TestHealth_UnhealthyServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	assert.False(t, health(server.Listener.Addr().String()), "Expected health check to fail for an unhealthy server")
}

func TestHealth_ServerTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	assert.False(t, health(server.Listener.Addr().String()), "Expected health check to fail due to server timeout")
}

func TestHealth_ErrorSendingRequest(t *testing.T) {
	assert.False(t, health("server:9999"), "Expected health check to fail due to sending request to unreal server")
}

func TestForward_SuccessfulRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rw := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = forward(server.Listener.Addr().String(), rw, req)
	assert.NoError(t, err, "Expected no error for successful request")
	assert.Equal(t, http.StatusOK, rw.Code, "Expected status OK for successful request")
}

func TestForward_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	rw := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = forward(server.Listener.Addr().String(), rw, req)
	assert.NotEqual(t, err, "Expected no error, because server successfuly responsed with some error in his work")
	assert.Equal(t, http.StatusInternalServerError, rw.Code, "Expected status Service Unavailable for unreal server")
}

func TestForward_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}))
	defer server.Close()

	rw := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = forward(server.Listener.Addr().String(), rw, req)
	assert.Error(t, err, "Expected error for request timeout")
	assert.Equal(t, http.StatusServiceUnavailable, rw.Code, "Expected status Service Unavailable for request timeout")
}

func TestForward_RequestError(t *testing.T) {
	invalidServerAddr := "invalid_address"

	rw := httptest.NewRecorder()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)

	err := forward(invalidServerAddr, rw, req)
	assert.Error(t, err, "Expected error for invalid server address")
	assert.Equal(t, http.StatusServiceUnavailable, rw.Code, "Expected status Service Unavailable for invalid server address")
}

func TestForward_CopyHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rw := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = forward(server.Listener.Addr().String(), rw, req)
	assert.NoError(t, err, "Expected no error for successful request")

	expectedHeaders := map[string]string{
		"Content-Type":    "application/json",
		"X-Custom-Header": "custom-value",
	}
	for key, value := range expectedHeaders {
		assert.Equal(t, value, rw.Header().Get(key), "Expected header %s to have value %s", key, value)
	}
}

func TestHash_EmptyString(t *testing.T) {
	expected := fnv.New32a().Sum32()

	result := hash("")
	assert.Equal(t, expected, result, "Expected hash of empty string to be %d", expected)
}

func TestHash_NonEmptyString(t *testing.T) {
	result := hash("hello")
	expected := fnv.New32a()
	expected.Write([]byte("hello"))
	expectedHash := expected.Sum32()

	assert.Equal(t, expectedHash, result, "Expected hash of 'hello' to be %d", expectedHash)
}

func TestHash_DifferentStrings(t *testing.T) {
	hash1 := hash("hello")
	hash2 := hash("world")

	assert.NotEqual(t, hash1, hash2, "Expected hash of 'hello' and 'world' to be different")
}