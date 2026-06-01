package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// slackChannel posts notifications to a Slack incoming webhook. Phone delivery
// rides the Slack mobile app, so no separate push provider is needed.
type slackChannel struct {
	webhookURL string
	client     *http.Client
}

func NewSlackChannel(webhookURL string) Channel {
	return slackChannel{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

func (slackChannel) Name() string { return "slack" }

func (s slackChannel) Dispatch(label, msg string) error {
	text := fmt.Sprintf("🐝 *%s*", label)
	if msg != "" {
		text += " — " + msg
	}
	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}
	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned %s", resp.Status)
	}
	return nil
}
