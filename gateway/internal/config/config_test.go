package config

import (
	"testing"
	"time"
)

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "defaults",
			want: Config{ListenAddr: ":8080", UpstreamURL: "https://api.anthropic.com", DisplayCurrency: "INR", ShutdownTimeout: 10 * time.Second},
		},
		{
			name: "overrides",
			env:  map[string]string{"TOKENGATE_LISTEN_ADDR": ":9000", "TOKENGATE_DISPLAY_CURRENCY": "USD", "TOKENGATE_SHUTDOWN_TIMEOUT": "3s"},
			want: Config{ListenAddr: ":9000", UpstreamURL: "https://api.anthropic.com", DisplayCurrency: "USD", ShutdownTimeout: 3 * time.Second},
		},
		{
			name:    "bad timeout",
			env:     map[string]string{"TOKENGATE_SHUTDOWN_TIMEOUT": "soon"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{"TOKENGATE_LISTEN_ADDR", "TOKENGATE_UPSTREAM_URL", "TOKENGATE_DISPLAY_CURRENCY", "TOKENGATE_SHUTDOWN_TIMEOUT"} {
				t.Setenv(k, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			got, err := FromEnv()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
