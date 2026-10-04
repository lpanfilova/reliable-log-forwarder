package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExampleLogEventEncodesAsJSON(t *testing.T) {
	event := exampleLogEvent()

	if event.Timestamp.IsZero() {
		t.Fatal("timestamp should be set")
	}
	if event.Level == "" {
		t.Fatal("level should be set")
	}
	if event.Message == "" {
		t.Fatal("message should be set")
	}

	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	var decoded struct {
		Timestamp time.Time `json:"timestamp"`
		Level     string    `json:"level"`
		Message   string    `json:"message"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	if !decoded.Timestamp.Equal(event.Timestamp) {
		t.Errorf("timestamp = %v, want %v", decoded.Timestamp, event.Timestamp)
	}
	if decoded.Level != event.Level {
		t.Errorf("level = %q, want %q", decoded.Level, event.Level)
	}
	if decoded.Message != event.Message {
		t.Errorf("message = %q, want %q", decoded.Message, event.Message)
	}
}

func TestForwardLogEvent(t *testing.T) {
	want := LogEvent{
		Timestamp: time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC),
		Level:     "info",
		Message:   "forwarded event",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/logs" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/logs")
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want %q", got, "application/json")
		}
		var got LogEvent
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if got != want {
			t.Errorf("event = %#v, want %#v", got, want)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := forwardLogEvent(want, server.URL); err != nil {
		t.Fatalf("forwardLogEvent returned error: %v", err)
	}
}

func TestMVPipeline(t *testing.T) {
	server := httptest.NewServer(logHandler())
	defer server.Close()

	event := exampleLogEvent()
	if err := forwardLogEvent(event, server.URL); err != nil {
		t.Fatalf("forwardLogEvent returned error: %v", err)
	}
}

func TestRunSend(t *testing.T) {
	server := httptest.NewServer(logHandler())
	defer server.Close()

	var stdout, stderr bytes.Buffer
	if err := run([]string{"send"}, &stdout, &stderr, server.URL); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "successfully") {
		t.Errorf("stdout = %q, want success message", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunInvalidModePrintsUsage(t *testing.T) {
	for _, args := range [][]string{nil, {"other"}} {
		var stdout, stderr bytes.Buffer
		if err := run(args, &stdout, &stderr, ""); err == nil {
			t.Errorf("run(%v) returned nil error, want error", args)
		}
		if !strings.Contains(stderr.String(), "usage:") {
			t.Errorf("run(%v) stderr = %q, want usage", args, stderr.String())
		}
	}
}

func TestForwardLogEventRejectsUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	err := forwardLogEvent(exampleLogEvent(), server.URL)
	if err == nil {
		t.Fatal("forwardLogEvent returned nil error for unexpected status")
	}
	if want := fmt.Sprintf("%d", http.StatusAccepted); !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want status code %s", err, want)
	}
}

func TestForwardLogEventReturnsRequestError(t *testing.T) {
	err := forwardLogEvent(exampleLogEvent(), "://invalid-url")
	if err == nil {
		t.Fatal("forwardLogEvent returned nil error for invalid receiver URL")
	}
}

func TestLogHandlerAcceptsValidJSON(t *testing.T) {
	body := `{"timestamp":"2026-10-03T12:00:00Z","level":"info","message":"hello"}`
	request := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(body))
	response := httptest.NewRecorder()

	logHandler().ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestLogHandlerRejectsInvalidJSON(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(`{"timestamp":`))
	response := httptest.NewRecorder()

	logHandler().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
