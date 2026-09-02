package errkind

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ChrisVerde02/ibmverify-go/client"

	"github.com/ChrisVerde02/ibmverify-cli/internal/exitcode"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, exitcode.OK},
		{"401", &client.APIError{StatusCode: 401}, exitcode.Auth},
		{"403", &client.APIError{StatusCode: 403}, exitcode.Auth},
		{"404", &client.APIError{StatusCode: 404}, exitcode.NotFound},
		{"429", &client.APIError{StatusCode: 429}, exitcode.RateLimit},
		{"500", &client.APIError{StatusCode: 500}, exitcode.Server},
		{"503", &client.APIError{StatusCode: 503}, exitcode.Server},
		{"wrapped 401", fmt.Errorf("op: %w", &client.APIError{StatusCode: 401}), exitcode.Auth},
		{"wrapped 404", fmt.Errorf("op: %w", &client.APIError{StatusCode: 404}), exitcode.NotFound},
		{"plain error", errors.New("something went wrong"), exitcode.Other},
		{"ErrNotFound sentinel", client.ErrNotFound, exitcode.NotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExitCode(tt.err)
			if got != tt.want {
				t.Errorf("ExitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
