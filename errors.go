package main

import "strings"

func isRateLimitErr(err error) bool {
	if err == nil {
		return false
	}
	return containsAnySubstrs(
		strings.ToLower(err.Error()),
		"ratelimit",
		"rate limit",
		"rate-limit",
		"rate_limit",
	)
}
