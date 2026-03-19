package aliyun

import (
	"math"
	"strings"
	"time"
)

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	retryableKeywords := []string{
		"timeout",
		"connection reset",
		"connection refused",
		"500",
		"502",
		"503",
		"504",
		"service unavailable",
		"throttling",
		"limit exceeded",
		"try again",
		"retry",
		"request timeout",
		"read timeout",
		"write timeout",
	}

	for _, keyword := range retryableKeywords {
		if strings.Contains(strings.ToLower(errStr), keyword) {
			return true
		}
	}

	return false
}

func backoff(attempt int) time.Duration {
	initialDelay := time.Second
	maxDelay := 30 * time.Second

	delay := time.Duration(math.Pow(2, float64(attempt))) * initialDelay

	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}
