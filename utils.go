package main

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func SafeLoadEnvs(filenames ...string) error {
	validFilenames := []string{}
	for _, fn := range filenames {
		if exists(fn) {
			validFilenames = append(validFilenames, fn)
		}
	}
	if len(validFilenames) == 0 {
		return nil
	}
	return godotenv.Overload(validFilenames...)
}

func isValidAuthToken(authToken string) bool {
	return len(authToken) >= 8
}

func isValidUrl(s string) bool {
	if s == "" {
		return false
	}
	_, err := url.Parse(s)
	return err == nil
}

func ensurePrefix(s, prefix string) string {
	if len(s) == 0 {
		return prefix
	}

	if strings.HasPrefix(s, prefix) {
		return s
	}

	return prefix + s
}

func remSuffixIfExists(s, suffix string) string {
	if len(s) == 0 {
		return ""
	}

	i := len(s) - len(suffix)
	if len(s) >= len(suffix) && s[i:] == suffix {
		return s[:i]
	}

	return s
}

func allCharsNumeric(s string) bool {
	for _, c := range s {
		_, err := strconv.Atoi(string(c))
		if err != nil {
			return false
		}
	}
	return true
}

func popStr(s *string) string {
	len := len(*s)
	if len == 0 {
		return ""
	}

	last := (*s)[len-1]
	*s = (*s)[:len-1]

	return string(last)
}
