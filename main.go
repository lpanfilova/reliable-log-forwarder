package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LogEvent is a single log entry.
type LogEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

func exampleLogEvent() LogEvent {
	return LogEvent{
		Timestamp: time.Now().UTC(),
		Level:     "info",
		Message:   "example log event",
	}
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, "http://localhost:8080"); err != nil {
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer, receiverURL string) error {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: go run . {receiver|send}")
		return fmt.Errorf("missing or invalid mode")
	}

	switch args[0] {
	case "receiver":
		if err := http.ListenAndServe(":8080", logHandler()); err != nil {
			return fmt.Errorf("start HTTP server: %w", err)
		}
	case "send":
		if err := forwardLogEvent(exampleLogEvent(), receiverURL); err != nil {
			fmt.Fprintln(stderr, "send log event:", err)
			return err
		}
		fmt.Fprintln(stdout, "log event delivered successfully")
	default:
		fmt.Fprintln(stderr, "usage: go run . {receiver|send}")
		return fmt.Errorf("unknown mode %q", args[0])
	}
	return nil
}

func logHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /logs", func(w http.ResponseWriter, r *http.Request) {
		var event LogEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "invalid log event", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return mux
}

func forwardLogEvent(event LogEvent, receiverURL string) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode log event: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, receiverURL+"/logs", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create log request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("send log event: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected receiver status: %s", response.Status)
	}
	return nil
}
