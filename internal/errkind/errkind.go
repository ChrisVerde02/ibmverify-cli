// Package errkind maps SDK errors to structured exit codes.
package errkind

import (
	"errors"

	"github.com/ChrisVerde02/ibmverify-go/client"

	"github.com/ChrisVerde02/ibmverify-cli/internal/exitcode"
)

// ExitCode inspects an error returned by the SDK and returns the appropriate
// exit code. It first checks for a typed *client.APIError (preferred), then
// falls back to string-scanning for errors not produced by the SDK.
func ExitCode(err error) int {
	if err == nil {
		return exitcode.OK
	}

	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.IsAuth():
			return exitcode.Auth
		case apiErr.IsNotFound():
			return exitcode.NotFound
		case apiErr.IsRateLimit():
			return exitcode.RateLimit
		case apiErr.IsServer():
			return exitcode.Server
		default:
			return exitcode.Other
		}
	}

	// Fallback: plain errors (flag validation, file I/O, etc.)
	return exitcode.Other
}
