package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/driftlog/internal/diff"
	"github.com/user/driftlog/internal/notify"
)

func driftedResults() []diff.DriftResult {
	return []diff.DriftResult{
		{ResourceID: "i-123", ResourceType: "aws_instance", Status: diff.StatusDrifted},
		{ResourceID: "bucket-a", ResourceType: "aws_s3_bucket", Status: diff.StatusOK},
	}
}

func TestSend_Webhook_Success(t *testing.T) {
	var received notify.Payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.New(notify.Config{
		Channel:    notify.ChannelWebhook,
		WebhookURL: ts.URL,
	})
	if err := n.Send(driftedResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.DriftCount != 1 {
		t.Errorf("expected drift_count=1, got %d", received.DriftCount)
	}
}

func TestSend_OnlyDrift_SkipsWhenNoDrift(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.New(notify.Config{
		Channel:    notify.ChannelWebhook,
		WebhookURL: ts.URL,
		OnlyDrift:  true,
	})
	cleanResults := []diff.DriftResult{
		{ResourceID: "i-ok", Status: diff.StatusOK},
	}
	if err := n.Send(cleanResults); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected webhook NOT to be called when no drift and only_drift=true")
	}
}

func TestSend_Slack_Success(t *testing.T) {
	var body map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body) //nolint
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.New(notify.Config{
		Channel:    notify.ChannelSlack,
		WebhookURL: ts.URL,
	})
	if err := n.Send(driftedResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body["text"] == "" {
		t.Error("expected non-empty slack text field")
	}
}

func TestSend_MissingURL_ReturnsError(t *testing.T) {
	n := notify.New(notify.Config{Channel: notify.ChannelWebhook})
	if err := n.Send(driftedResults()); err == nil {
		t.Error("expected error for missing webhook_url")
	}
}

func TestSend_UnknownChannel_ReturnsError(t *testing.T) {
	n := notify.New(notify.Config{
		Channel:    "pagerduty",
		WebhookURL: "http://example.com",
	})
	if err := n.Send(driftedResults()); err == nil {
		t.Error("expected error for unknown channel")
	}
}

func TestSend_ServerError_ReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notify.New(notify.Config{Channel: notify.ChannelWebhook, WebhookURL: ts.URL})
	if err := n.Send(driftedResults()); err == nil {
		t.Error("expected error on 500 response")
	}
}
