package notifications

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSlackChannelDispatch(t *testing.T) {
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ch := NewSlackChannel(srv.URL)
	if err := ch.Dispatch("alpha", "tests passing"); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	text := gotBody["text"]
	if !strings.Contains(text, "alpha") || !strings.Contains(text, "tests passing") {
		t.Errorf("posted text = %q, want it to mention label and message", text)
	}
}

func TestSlackChannelDispatchNoMessage(t *testing.T) {
	var gotText string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		raw, _ := io.ReadAll(r.Body)
		json.Unmarshal(raw, &body)
		gotText = body["text"]
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := NewSlackChannel(srv.URL).Dispatch("beta", ""); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !strings.Contains(gotText, "beta") || strings.Contains(gotText, "—") {
		t.Errorf("posted text = %q, want just the label with no separator", gotText)
	}
}

func TestSlackChannelDispatchErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	if err := NewSlackChannel(srv.URL).Dispatch("alpha", "msg"); err == nil {
		t.Error("expected an error on non-2xx response, got nil")
	}
}
