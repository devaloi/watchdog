package action

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devaloi/watchdog/internal/watcher"
)

func TestWebhookActionExecute(t *testing.T) {
	var received WebhookPayload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			t.Fatal(readErr)
		}

		decodeErr := json.Unmarshal(body, &received)
		if decodeErr != nil {
			t.Fatal(decodeErr)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := NewWebhookAction(srv.URL, http.MethodPost, nil, 5*time.Second)

	ev := watcher.Event{Path: "main.go", Type: watcher.Modify, Name: "main.go", Dir: "."}

	err := wh.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	if received.Path != "main.go" {
		t.Errorf("payload path = %q, want %q", received.Path, "main.go")
	}

	if received.Event != "modify" {
		t.Errorf("payload event = %q, want %q", received.Event, "modify")
	}
}

func TestWebhookActionCustomHeaders(t *testing.T) {
	var gotHeader string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Custom")

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	headers := map[string]string{"X-Custom": "test-value"}
	wh := NewWebhookAction(srv.URL, http.MethodPost, headers, 5*time.Second)

	ev := watcher.Event{Path: "a.go", Type: watcher.Create, Name: "a.go", Dir: "."}

	err := wh.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	if gotHeader != "test-value" {
		t.Errorf("custom header = %q, want %q", gotHeader, "test-value")
	}
}

func TestWebhookActionDryRun(t *testing.T) {
	called := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := NewWebhookAction(srv.URL, http.MethodPost, nil, 5*time.Second)
	wh.DryRun = true

	ev := watcher.Event{Path: "a.go", Type: watcher.Create, Name: "a.go", Dir: "."}

	err := wh.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	if called {
		t.Error("webhook should not be called in dry run mode")
	}
}
