package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishTweetOpts(t *testing.T) {
	t.Run("Test handleFetchJsonResp()", func(t *testing.T) {
		type PublishTweetOptsTest struct {
			text      string
			expected  string
			shouldErr bool
		}

		var newPublishTweetOptsTest = func(text, expected string, shouldErr bool) *PublishTweetOptsTest {
			return &PublishTweetOptsTest{
				text:      text,
				expected:  expected,
				shouldErr: shouldErr,
			}
		}

		jsonBytes := []byte(`{
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

		tests := []*PublishTweetOptsTest{
			// Empty string
			newPublishTweetOptsTest("", string(jsonBytes), false),

			// No substitutions
			newPublishTweetOptsTest("some text", "some text", false),

			// Non-existing fields
			newPublishTweetOptsTest("{*{ stuff }*}", "", true),
			newPublishTweetOptsTest("{*{ data.stuff }*}", "", true),

			// Post fields
			newPublishTweetOptsTest("{*{ data.post.title }*}", "My Awesome Title", false),
			newPublishTweetOptsTest("{*{ data.post.timestamp }*}", "12345678", false),
			newPublishTweetOptsTest("{*{ data.post.foo }*}", "true", false),
			newPublishTweetOptsTest("{*{ data.post.bar }*}", "false", false),
			newPublishTweetOptsTest("{*{ data.post.tags }*}", `["Awesome","Cool"]`, false),

			// Nested object
			newPublishTweetOptsTest("{*{ data.post.stats.status }*}", "published", false),
			newPublishTweetOptsTest("{*{ data.post.stats.traffic }*}", "87654321", false),
			newPublishTweetOptsTest("{*{ data.post.stats.trending }*}", "false", false),
			newPublishTweetOptsTest("{*{ data.post.stats.hot }*}", "true", false),
			newPublishTweetOptsTest("{*{ data.post.stats.more }*}", `{"some":"data"}`, false),
			newPublishTweetOptsTest("{*{ data.post.stats.more.some }*}", "data", false),
			newPublishTweetOptsTest("{*{ data.post.stats.items }*}", `["foo","bar","baz"]`, false),

			// Indexing inside of array (TODO: not yet implimented)
			newPublishTweetOptsTest("{*{ data.post.stats.items[0] }*}", "", true),
			newPublishTweetOptsTest("{*{ data.post.stats.items[1] }*}", "", true),
			newPublishTweetOptsTest("{*{ data.post.stats.items[2] }*}", "", true),

			// Objects inside of array
			newPublishTweetOptsTest("{*{ data.post.authors }*}", `[{"name":"Jim"},{"name":"Bob"}]`, false),

			// Indexing objects inside of array (TODO: not yet implimented)
			newPublishTweetOptsTest("{*{ data.post.authors[0] }*}", "", true),
			newPublishTweetOptsTest("{*{ data.post.authors[1] }*}", "", true),
		}

		for _, test := range tests {
			opts := PublishTweetOpts{
				Text: test.text,
			}

			// A new *http.Response is needed for each test,
			// because handleFetchJsonResp() clears the response body
			resp := &http.Response{
				Body: NewByteReadCloser(jsonBytes),
			}

			s, err := opts.handleFetchJsonResp(resp)
			assert.Equal(t, test.expected, s)
			if test.shouldErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	})

	t.Run("Test getReplyToTweetID()", func(t *testing.T) {
		type GetReplyToTweetIDTest struct {
			replyTo   string
			expected  string
			shouldErr bool
		}

		tests := []GetReplyToTweetIDTest{
			// Empty string
			{
				replyTo:   "",
				expected:  "",
				shouldErr: true,
			},

			// ReplyTo Tweet id
			{
				replyTo:   "1234567890123456789",
				expected:  "1234567890123456789",
				shouldErr: false,
			},

			// ReplyTo Tweet urls
			{
				replyTo:   "https://twitter.com/user/status/1234567890123456789",
				expected:  "1234567890123456789",
				shouldErr: false,
			},
			{
				replyTo:   "https://twitter.com/user/status/1234567890123456789/",
				expected:  "1234567890123456789",
				shouldErr: false,
			},
			{
				replyTo:   "https://twitter.com/user/status/1234567890123456789/?",
				expected:  "1234567890123456789",
				shouldErr: false,
			},
			{
				replyTo:   "https://twitter.com/user/status/1234567890123456789?",
				expected:  "1234567890123456789",
				shouldErr: false,
			},
			{
				replyTo:   "https://twitter.com/user/status/1234567890123456789?=",
				expected:  "1234567890123456789",
				shouldErr: false,
			},
			{
				replyTo:   "https://twitter.com/user/status/1234567890123456789?qa=",
				expected:  "1234567890123456789",
				shouldErr: false,
			},
			{
				replyTo:   "https://twitter.com/user/status/1234567890123456789?a=123",
				expected:  "1234567890123456789",
				shouldErr: false,
			},
			{
				replyTo:   "https://twitter.com/user/status/1234567890123456789?a=123&b=456",
				expected:  "1234567890123456789",
				shouldErr: false,
			},
			{
				replyTo:   "https://twitter.com/user/status/1234567890123456789#",
				expected:  "1234567890123456789",
				shouldErr: false,
			},
			{
				replyTo:   "https://twitter.com/user/status/1234567890123456789#abc",
				expected:  "1234567890123456789",
				shouldErr: false,
			},
		}

		for _, test := range tests {
			opts := PublishTweetOpts{
				ReplyTo: test.replyTo,
			}

			tweetID, err := opts.getReplyToTweetID()
			assert.Equal(t, test.expected, tweetID)
			if test.shouldErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	})

	t.Run("Test validReplyTo()", func(t *testing.T) {
		type ValidReplyToTest struct {
			replyTo  string
			expected bool
		}

		tests := []ValidReplyToTest{
			// Empty string
			{
				replyTo:  "",
				expected: false,
			},

			// Proper usage
			{
				replyTo:  "1234567890123456789",
				expected: true,
			},
			{
				replyTo:  "https://twitter.com/user/status/1234567890123456789",
				expected: true,
			},

			// Invalid Tweet ID
			{
				replyTo:  "https://twitter.com/user/status/123",
				expected: false,
			},
			{
				replyTo:  "https://twitter.com/user/status/abc",
				expected: false,
			},
			{
				replyTo:  "https://twitter.com/user/status/abcdefghijklmnopqrs",
				expected: false,
			},
		}

		for _, test := range tests {
			opts := PublishTweetOpts{
				ReplyTo: test.replyTo,
			}

			assert.Equal(t, test.expected, opts.validReplyTo())
		}
	})

	t.Run("Test jsonFmts()", func(t *testing.T) {
		type JsonFmtsTest struct {
			text     string
			expected []string
		}

		tests := []JsonFmtsTest{
			{
				text:     "The article title is {*{ data.post.title }*}.",
				expected: []string{"data.post.title"},
			},
			{
				text:     "The article title is {*{ data.post.title }*}, and the timestamp is {*{ data.post.timestamp }*}.",
				expected: []string{"data.post.title", "data.post.timestamp"},
			},
		}

		for _, test := range tests {
			opts := PublishTweetOpts{
				Text: test.text,
			}

			assert.Equal(t, test.expected, opts.JsonFmts())
		}
	})
}
