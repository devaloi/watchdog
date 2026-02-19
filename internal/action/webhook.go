package action

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/devaloi/watchdog/internal/watcher"
)

const defaultWebhookTimeout = 10 * time.Second

// WebhookPayload is the JSON body sent to webhook URLs.
type WebhookPayload struct {
	Path  string `json:"path"`
	Event string `json:"event"`
	Time  string `json:"time"`
}

// WebhookAction sends an HTTP request when triggered.
type WebhookAction struct {
	URL     string
	Method  string
	Headers map[string]string
	Timeout time.Duration
	DryRun  bool
	client  *http.Client
}

// NewWebhookAction creates a WebhookAction with the given URL and method.
func NewWebhookAction(url, method string, headers map[string]string, timeout time.Duration) *WebhookAction {
	if method == "" {
		method = http.MethodPost
	}

	if timeout == 0 {
		timeout = defaultWebhookTimeout
	}

	return &WebhookAction{
		URL:     url,
		Method:  method,
		Headers: headers,
		Timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// Execute sends the HTTP request with event data as JSON.
func (w *WebhookAction) Execute(ev watcher.Event) error {
	if w.DryRun {
		return nil
	}

	payload := WebhookPayload{
		Path:  ev.Path,
		Event: string(ev.Type),
		Time:  time.Now().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		w.Method,
		w.URL,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range w.Headers {
		req.Header.Set(k, v)
	}

	resp, err := w.client.Do(req) //nolint:gosec // URL is user-configured by design
	if err != nil {
		return err
	}

	return resp.Body.Close()
}
