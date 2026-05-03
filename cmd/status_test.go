package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// emptyData represents a queue with no waiting sessions.
func emptyData() *statusData {
	return &statusData{Count: 0, Queue: []statusEntry{}}
}

// singleData represents a queue with one waiting session.
func singleData() *statusData {
	q := []statusEntry{{
		Session:    "alpha",
		Message:    "tests passing",
		Pane:       "%5",
		Age:        "2m ago",
		NotifiedAt: time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC),
	}}
	return &statusData{
		Count: 1,
		Next:  &q[0],
		Queue: q,
	}
}

// multiData represents three sessions with `alpha` at the head.
func multiData() *statusData {
	q := []statusEntry{
		{Session: "alpha", Message: "tests passing", Pane: "%5", Age: "5m ago"},
		{Session: "beta", Message: "blocked", Pane: "%6", Age: "3m ago"},
		{Session: "gamma", Message: "", Pane: "%7", Age: "1m ago"},
	}
	return &statusData{
		Count: 3,
		Next:  &q[0],
		Queue: q,
	}
}

func TestExecTemplate_HumanEmpty(t *testing.T) {
	tmpl := `{{- if eq .Count 0 -}}No sessions waiting{{- else -}}{{ .Count }} session(s) waiting{{- end -}}`
	var buf bytes.Buffer
	if err := execTemplate(&buf, tmpl, emptyData()); err != nil {
		t.Fatalf("execTemplate: %v", err)
	}
	if got := buf.String(); got != "No sessions waiting" {
		t.Errorf("got %q, want %q", got, "No sessions waiting")
	}
}

func TestExecTemplate_HumanSingle(t *testing.T) {
	tmpl := `{{ .Count }} session(s); next: {{ .Next.Session }} ({{ .Next.Age }})`
	var buf bytes.Buffer
	if err := execTemplate(&buf, tmpl, singleData()); err != nil {
		t.Fatalf("execTemplate: %v", err)
	}
	want := "1 session(s); next: alpha (2m ago)"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExecTemplate_HumanMessageCollapsesWhenEmpty(t *testing.T) {
	// gamma has empty message — the conditional should collapse the surrounding " — ".
	tmpl := `next: {{ .Next.Session }}{{ if .Next.Message }} — {{ .Next.Message }}{{ end }}`
	data := &statusData{
		Count: 1,
		Next:  &statusEntry{Session: "gamma", Message: ""},
	}
	var buf bytes.Buffer
	if err := execTemplate(&buf, tmpl, data); err != nil {
		t.Fatalf("execTemplate: %v", err)
	}
	want := "next: gamma"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q (collapse fail)", got, want)
	}
}

func TestExecTemplate_TmuxMulti(t *testing.T) {
	tmpl := `{{- if .Next -}}🐝 {{ .Next.Session }}{{ if gt .Count 1 }} | +{{ len (slice .Queue 1) }}{{ end }}{{ end -}}`
	var buf bytes.Buffer
	if err := execTemplate(&buf, tmpl, multiData()); err != nil {
		t.Fatalf("execTemplate: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "🐝 alpha") {
		t.Errorf("expected 🐝 alpha in output, got %q", got)
	}
	if !strings.Contains(got, "+2") {
		t.Errorf("expected +2 (extra count) in output, got %q", got)
	}
}

func TestExecTemplate_TmuxEmpty(t *testing.T) {
	// Tmux template produces no output when the queue is empty.
	tmpl := `{{- if .Next -}}🐝 {{ .Next.Session }}{{ end -}}`
	var buf bytes.Buffer
	if err := execTemplate(&buf, tmpl, emptyData()); err != nil {
		t.Fatalf("execTemplate: %v", err)
	}
	if got := buf.String(); got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestExecTemplate_RejectsBadTemplate(t *testing.T) {
	if err := execTemplate(&bytes.Buffer{}, `{{ .Bad`, singleData()); err == nil {
		t.Error("expected parse error for malformed template, got nil")
	}
}

func TestStatusJSON_EmptyShape(t *testing.T) {
	data := emptyData()
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	queue, ok := got["queue"].([]any)
	if !ok {
		t.Fatalf("queue is not a list: %T", got["queue"])
	}
	if len(queue) != 0 {
		t.Errorf("queue len = %d, want 0", len(queue))
	}
	// Derivable fields must not be in the JSON output — consumers compute
	// them from queue.
	for _, key := range []string{"count", "next", "extra"} {
		if _, present := got[key]; present {
			t.Errorf("derivable field %q must not be in JSON output", key)
		}
	}
}

func TestStatusJSON_PopulatedShape(t *testing.T) {
	data := multiData()
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	queue, ok := got["queue"].([]any)
	if !ok {
		t.Fatalf("queue is not a list: %T", got["queue"])
	}
	if len(queue) != 3 {
		t.Errorf("queue len = %d, want 3", len(queue))
	}
	first := queue[0].(map[string]any)
	if first["session"] != "alpha" {
		t.Errorf("queue[0].session = %v, want alpha", first["session"])
	}
	for _, key := range []string{"count", "next", "extra"} {
		if _, present := got[key]; present {
			t.Errorf("derivable field %q must not be in JSON output", key)
		}
	}
}
