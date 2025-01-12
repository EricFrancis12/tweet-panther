package main

import "bytes"

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

type PublishTweetType string

const (
	PublishTweetTypeText      PublishTweetType = "text"
	PublishTweetTypeFetchJson PublishTweetType = "fetch_json"
)

const ContentTypeApplicationJson string = "application/json"

const (
	EnvPort             string = "PORT"
	EnvAuthToken        string = "AUTH_TOKEN"
	EnvAPIKey           string = "API_KEY"
	EnvAPIKeySecret     string = "API_KEY_SECRET"
	EnvOAuthToken       string = "O_AUTH_TOKEN"
	EnvOAuthTokenSecret string = "O_AUTH_TOKEN_SECRET"
)

const (
	HTTPHeaderAuthorization string = "Authorization"
	HTTPHeaderContentType   string = "Content-Type"
)
