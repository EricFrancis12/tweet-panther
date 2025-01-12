package main

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ByteReadCloser struct {
	reader *bytes.Reader
}

func NewByteReadCloser(data []byte) *ByteReadCloser {
	return &ByteReadCloser{
		reader: bytes.NewReader(data),
	}
}

func (b *ByteReadCloser) Read(p []byte) (n int, err error) {
	return b.reader.Read(p)
}

func (b *ByteReadCloser) Close() error {
	b.reader = nil
	return nil
}

type JsonFmtTest struct {
	jsonFmt  string
	expected string
}

func newJsonFmtTest(jsonFmt, expected string) *JsonFmtTest {
	return &JsonFmtTest{
		jsonFmt:  jsonFmt,
		expected: expected,
	}
}

func TestHandleFetchJsonResp(t *testing.T) {
	resp := &http.Response{
		Body: NewByteReadCloser([]byte(`{
			"data": {
				"post": {
					"title": "My Awesome Title",
					"timestamp": 12345678
				}
			}
		}`)),
	}

	jsonFmtTests := []*JsonFmtTest{
		newJsonFmtTest("data.post.title", "My Awesome Title"),
		// TODO: newJsonFmtTest("data.post.timestamp", "12345678"),
	}

	for _, j := range jsonFmtTests {
		opts := PublishTweetOpts{
			JsonFmt: j.jsonFmt,
		}

		s, err := opts.handleFetchJsonResp(resp)
		assert.Nil(t, err)
		assert.Equal(t, j.expected, s)
	}
}
