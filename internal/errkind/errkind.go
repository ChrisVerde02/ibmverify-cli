// Package errkind maps SDK error messages to structured exit codes.
package errkind

import (
	"strings"

	"github.com/ChrisVerde02/ibmverify-cli/internal/exitcode"
)

// ExitCode inspects an error returned by the SDK and returns the appropriate
// exit code. It works by scanning the error string for known HTTP status
// phrases since the SDK embeds the status code in the message.
func ExitCode(err error) int {
	if err == nil {
		return exitcode.OK
	}
	msg := err.Error()
	switch {
	case contains(msg, "HTTP 401", "HTTP 403", "invalid_client", "unauthorized"):
		return exitcode.Auth
	case contains(msg, "HTTP 404", "not found"):
		return exitcode.NotFound
	case contains(msg, "HTTP 429", "rate"):
		return exitcode.RateLimit
	case contains(msg, "HTTP 5"):
		return exitcode.Server
	case contains(msg, "cannot be empty", "invalid"):
		return exitcode.Validation
	default:
		return exitcode.Other
	}
}

func contains(s string, subs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range subs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}
