package main

import (
	"fmt"
	"strings"
)

func errInvalidTweetID(tweetID string) error {
	return fmt.Errorf("tweet ID must be 19 characters long, and contain only numeric characters (received: %s)", tweetID)
}

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
