package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/driftlog/internal/diff"
)

// Channel represents a notification destination type.
type Channel string

const (
	ChannelSlack   Channel = "slack"
	ChannelWebhook Channel = "webhook"
)

// Config holds notification settings.
type Config struct {
	Channel    Channel `yaml:"channel"`
	WebhookURL string  `yaml:"webhook_url"`
	OnlyDrift  bool    `yaml:"only_drift"`
}

// Payload is the JSON body sent to a webhook.
type Payload struct {
	Timestamp  time.Time        `json:"timestamp"`
	DriftCount int              `json:"drift_count"`
	Results    []diff.DriftResult `json:"results"`
}

// Notifier sends drift results to a configured channel.
type Notifier struct {
	cfg    Config
	client *http.Client
}

// New creates a Notifier with the given config.
func New(cfg Config) *Notifier {
	return &Notifier{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send dispatches results according to the configured channel.
// If OnlyDrift is true, it skips sending when there are no drifted resources.
func (n *Notifier) Send(results []diff.DriftResult) error {
	if n.cfg.WebhookURL == "" {
		return fmt.Errorf("notify: webhook_url is required")
	}

	drifted := countDrifted(results)
	if n.cfg.OnlyDrift && drifted == 0 {
		return nil
	}

	payload := Payload{
		Timestamp:  time.Now().UTC(),
		DriftCount: drifted,
		Results:    results,
	}

	switch n.cfg.Channel {
	case ChannelSlack:
		return n.sendSlack(payload)
	case ChannelWebhook, "":
		return n.sendWebhook(payload)
	default:
		return fmt.Errorf("notify: unknown channel %q", n.cfg.Channel)
	}
}

func (n *Notifier) sendWebhook(p Payload) error {
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}
	resp, err := n.client.Post(n.cfg.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: http post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notify: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (n *Notifier) sendSlack(p Payload) error {
	text := fmt.Sprintf(":warning: *DriftLog*: %d drifted resource(s) detected at %s",
		p.DriftCount, p.Timestamp.Format(time.RFC3339))
	slackBody, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("notify: marshal slack payload: %w", err)
	}
	resp, err := n.client.Post(n.cfg.WebhookURL, "application/json", bytes.NewReader(slackBody))
	if err != nil {
		return fmt.Errorf("notify: http post slack: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notify: slack unexpected status %d", resp.StatusCode)
	}
	return nil
}

func countDrifted(results []diff.DriftResult) int {
	n := 0
	for _, r := range results {
		if r.Status == diff.StatusDrifted || r.Status == diff.StatusMissing || r.Status == diff.StatusExtra {
			n++
		}
	}
	return n
}
