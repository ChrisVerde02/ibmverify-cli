// Package exitcode defines the standard exit codes for ibmverify-cli.
//
//	0 — success
//	1 — other / unknown error
//	2 — usage / flag validation error
//	3 — authentication failure (401/403)
//	4 — resource not found (404)
//	5 — rate limited (429)
//	6 — server error (5xx)
//	7 — validation error (bad input before any HTTP call)
package exitcode

const (
	OK         = 0
	Other      = 1
	Usage      = 2
	Auth       = 3
	NotFound   = 4
	RateLimit  = 5
	Server     = 6
	Validation = 7
)
