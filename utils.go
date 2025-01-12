package main

import (
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func SafeLoadEnvs(filenames ...string) error {
	validFilenames := []string{}
	for _, fn := range filenames {
		if fileExists(fn) {
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

	if string(s[0]) == prefix {
		return s
	}

	return prefix + s
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
