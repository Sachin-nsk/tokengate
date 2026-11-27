// Command fakeupstream serves an Anthropic-compatible mock for CI and load tests.
package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/sachin-nsk/tokengate/gateway/internal/fakeupstream"
)

func main() {
	addr := flag.String("addr", ":9090", "listen address")
	latency := flag.Duration("latency", 0, "delay before responding")
	in := flag.Int("input-tokens", 100, "reported input tokens")
	out := flag.Int("output-tokens", 50, "reported output tokens")
	errRate := flag.Float64("error-rate", 0, "fraction of requests failing with 529")
	flag.Parse()

	srv := &http.Server{
		Addr:              *addr,
		Handler:           fakeupstream.Handler(fakeupstream.Options{Latency: *latency, InputTokens: *in, OutputTokens: *out, ErrorRate: *errRate}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("fake upstream started", "addr", *addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
