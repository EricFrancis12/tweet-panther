package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type JsonFmtTest struct {
	jsonFmt   string
	expected  string
	shouldErr bool
}

func newJsonFmtTest(jsonFmt, expected string, shouldErr bool) *JsonFmtTest {
	return &JsonFmtTest{
		jsonFmt:   jsonFmt,
		expected:  expected,
		shouldErr: shouldErr,
	}
}

func TestHandleFetchJsonResp(t *testing.T) {
	jsonStr := []byte(`{
		"data": {
			"post": {
				"title": "My Awesome Title",
				"timestamp": 12345678,
				"foo": true,
				"bar": false,
				"tags": [
					"Awesome",
					"Cool"
				],
				"stats": {
					"status": "published",
					"traffic": 87654321,
					"trending": false,
					"hot": true,
					"more": {
						"some": "data"
					},
					"items": [
						"foo",
						"bar",
						"baz"
					]
				},
				"authors": [
					{
						"name": "Jim"
					},
					{
						"name": "Bob"
					}
				]
			}
		}
	}`)

	jsonFmtTests := []*JsonFmtTest{
		// Empty string
		newJsonFmtTest("", string(jsonStr), false),

		// Non-existing field
		newJsonFmtTest("stuff", "", true),

		// Post fields
		newJsonFmtTest("data.post.title", "My Awesome Title", false),
		newJsonFmtTest("data.post.timestamp", "12345678", false),
		newJsonFmtTest("data.post.foo", "true", false),
		newJsonFmtTest("data.post.bar", "false", false),
		newJsonFmtTest("data.post.tags", `["Awesome","Cool"]`, false),

		// Nested object
		newJsonFmtTest("data.post.stats.status", "published", false),
		newJsonFmtTest("data.post.stats.traffic", "87654321", false),
		newJsonFmtTest("data.post.stats.trending", "false", false),
		newJsonFmtTest("data.post.stats.hot", "true", false),
		newJsonFmtTest("data.post.stats.more", `{"some":"data"}`, false),
		newJsonFmtTest("data.post.stats.more.some", "data", false),
		newJsonFmtTest("data.post.stats.items", `["foo","bar","baz"]`, false),

		// Indexing inside of array (TODO: not yet implimented)
		newJsonFmtTest("data.post.stats.items[0]", "", true),
		newJsonFmtTest("data.post.stats.items[1]", "", true),
		newJsonFmtTest("data.post.stats.items[2]", "", true),

		// Objects inside of array
		newJsonFmtTest("data.post.authors", `[{"name":"Jim"},{"name":"Bob"}]`, false),

		// Indexing objects inside of array (TODO: not yet implimented)
		newJsonFmtTest("data.post.authors[0]", "", true),
		newJsonFmtTest("data.post.authors[1]", "", true),
	}

	for _, j := range jsonFmtTests {
		opts := PublishTweetOpts{
			JsonFmt: j.jsonFmt,
		}

		// A new *http.Response is needed for each j,
		// because handleFetchJsonResp() clears the response body
		resp := &http.Response{
			Body: NewByteReadCloser(jsonStr),
		}

		s, err := opts.handleFetchJsonResp(resp)
		assert.Equal(t, j.expected, s)
		if j.shouldErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}
