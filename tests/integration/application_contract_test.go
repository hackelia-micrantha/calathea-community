//go:build integration

package integration_test

import (
	"bytes"
	"testing"

	"github.com/hackelia-micrantha/calathea-community/internal/application"
)

func TestApplicationCommandContract(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "default command",
			wantCode:   0,
			wantStdout: "calathea: local portfolio orientation\n",
		},
		{
			name:       "version",
			args:       []string{"version"},
			wantCode:   0,
			wantStdout: "calathea dev\n",
		},
		{
			name:       "unknown command",
			args:       []string{"unknown"},
			wantCode:   2,
			wantStderr: "unknown command: unknown\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := application.Run(&stdout, &stderr, tt.args)
			if code != tt.wantCode {
				t.Fatalf("Run() code = %d, want %d", code, tt.wantCode)
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}
