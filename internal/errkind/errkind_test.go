package errkind

import (
	"errors"
	"testing"

	"github.com/ChrisVerde02/ibmverify-cli/internal/exitcode"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, exitcode.OK},
		{"401", errors.New("IBM Verify failed with HTTP 401: unauthorized"), exitcode.Auth},
		{"403", errors.New("IBM Verify failed with HTTP 403: forbidden"), exitcode.Auth},
		{"invalid_client", errors.New("invalid_client: bad credentials"), exitcode.Auth},
		{"404", errors.New("IBM Verify failed with HTTP 404: not found"), exitcode.NotFound},
		{"429", errors.New("IBM Verify failed with HTTP 429: rate limit"), exitcode.RateLimit},
		{"500", errors.New("IBM Verify failed with HTTP 500: internal server error"), exitcode.Server},
		{"empty field", errors.New("client ID cannot be empty"), exitcode.Validation},
		{"other", errors.New("something unexpected"), exitcode.Other},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExitCode(tt.err)
			if got != tt.want {
				t.Errorf("ExitCode(%q) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
