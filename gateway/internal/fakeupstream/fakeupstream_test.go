package fakeupstream

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func post(t *testing.T, h http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body)))
	return rec
}

func TestNonStreaming(t *testing.T) {
	rec := post(t, Handler(Options{InputTokens: 10, OutputTokens: 5}), `{"model":"claude-haiku-5-5"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var got struct {
		Model string `json:"model"`
		Usage struct {
			In  int `json:"input_tokens"`
			Out int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Model != "claude-haiku-5-5" || got.Usage.In != 10 || got.Usage.Out != 5 {
		t.Fatalf("unexpected response %+v", got)
	}
}

func TestStreamingEventOrder(t *testing.T) {
	rec := post(t, Handler(Options{OutputTokens: 5}), `{"model":"m","stream":true}`)
	body := rec.Body.String()
	last := -1
	for _, ev := range []string{"message_start", "content_block_start", "content_block_delta", "content_block_stop", "message_delta", "message_stop"} {
		i := strings.Index(body, "event: "+ev)
		if i <= last {
			t.Fatalf("event %s missing or out of order in:\n%s", ev, body)
		}
		last = i
	}
}

func TestErrorRate(t *testing.T) {
	rec := post(t, Handler(Options{ErrorRate: 1}), `{"model":"m"}`)
	if rec.Code != 529 {
		t.Fatalf("status %d, want 529", rec.Code)
	}
}

func TestBadJSON(t *testing.T) {
	if rec := post(t, Handler(Options{}), `{`); rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}
