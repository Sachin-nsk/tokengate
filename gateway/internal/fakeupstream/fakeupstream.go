// Package fakeupstream is an Anthropic Messages API mock used by CI and load tests.
package fakeupstream

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

// Options configure the mock's behaviour.
type Options struct {
	Latency      time.Duration // delay before the first byte
	InputTokens  int           // reported usage.input_tokens
	OutputTokens int           // reported usage.output_tokens
	ErrorRate    float64       // fraction of requests answered with HTTP 529
	ReplyText    string
}

// Handler returns an http.Handler serving POST /v1/messages.
func Handler(o Options) http.Handler {
	if o.ReplyText == "" {
		o.ReplyText = "ok"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/messages", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Model  string `json:"model"`
			Stream bool   `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request_error", "invalid JSON body")
			return
		}
		select {
		case <-time.After(o.Latency):
		case <-r.Context().Done():
			return
		}
		if o.ErrorRate > 0 && rand.Float64() < o.ErrorRate { //nolint:gosec // test mock, not security-sensitive
			writeError(w, 529, "overloaded_error", "overloaded")
			return
		}
		if req.Stream {
			stream(w, req.Model, o)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(message(req.Model, o))
	})
	return mux
}

func message(model string, o Options) map[string]any {
	return map[string]any{
		"id":            "msg_fake",
		"type":          "message",
		"role":          "assistant",
		"model":         model,
		"content":       []map[string]any{{"type": "text", "text": o.ReplyText}},
		"stop_reason":   "end_turn",
		"stop_sequence": nil,
		"usage": map[string]any{
			"input_tokens":                o.InputTokens,
			"output_tokens":               o.OutputTokens,
			"cache_creation_input_tokens": 0,
			"cache_read_input_tokens":     0,
		},
	}
}

func stream(w http.ResponseWriter, model string, o Options) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	f, _ := w.(http.Flusher)
	send := func(event string, data any) {
		b, _ := json.Marshal(data)
		_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
		if f != nil {
			f.Flush()
		}
	}
	start := message(model, o)
	start["content"] = []any{}
	start["stop_reason"] = nil
	start["usage"] = map[string]any{"input_tokens": o.InputTokens, "output_tokens": 1}
	send("message_start", map[string]any{"type": "message_start", "message": start})
	send("content_block_start", map[string]any{"type": "content_block_start", "index": 0, "content_block": map[string]any{"type": "text", "text": ""}})
	send("content_block_delta", map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "text_delta", "text": o.ReplyText}})
	send("content_block_stop", map[string]any{"type": "content_block_stop", "index": 0})
	send("message_delta", map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": "end_turn", "stop_sequence": nil}, "usage": map[string]any{"output_tokens": o.OutputTokens}})
	send("message_stop", map[string]any{"type": "message_stop"})
}

func writeError(w http.ResponseWriter, status int, typ, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":  "error",
		"error": map[string]any{"type": typ, "message": msg},
	})
}
