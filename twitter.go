package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/EricFrancis12/stripol"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	managetweetTypes "github.com/michimani/gotwi/tweet/managetweet/types"
)

type TwitterAPICreds struct {
	APIKey           string
	APIKeySecret     string
	OAuthToken       string
	OAuthTokenSecret string
}

func (tac TwitterAPICreds) isValid() bool {
	return tac.APIKey != "" && tac.APIKeySecret != "" && tac.OAuthToken != "" && tac.OAuthTokenSecret != ""
}

type PublishTweetOpts struct {
	PublishTweetType PublishTweetType `json:"publishTweetType"`
	Text             string           `json:"text"`
	ReplyTo          string           `json:"replyTo"`
	Url              string           `json:"url"`
}

func (o PublishTweetOpts) handleFetchJsonResp(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if o.Text == "" {
		return string(body), nil
	}

	var data interface{}
	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	if err := d.Decode(&data); err != nil {
		return "", err
	}

	var (
		text = o.Text
		ipol = stripol.New("{{", "}}")
	)

	for _, jsonFmt := range o.JsonFmts() {
		keys := strings.Split(jsonFmt, ".")
		for _, key := range keys {
			m, ok := data.(map[string]interface{})
			if !ok {
				return "", errors.New("invalid jsonFmt")
			}

			if data, ok = m[key]; !ok {
				return "", errors.New("invalid jsonFmt")
			}
		}

		var s string

		switch d := data.(type) {
		case string:
			s = d
		case int64:
			s = fmt.Sprintf("%d", d)
		case float64:
			s = fmt.Sprintf("%f", d)
		case bool:
			s = strconv.FormatBool(d)
		default:
			b, err := json.Marshal(data)
			if err != nil {
				return "", err
			}
			s = string(b)
		}

		ipol.RegisterVar(jsonFmt, s)
	}

	f := newFuncIpol("|*", "*|")
	f.RegisterFn("pathEscape", func(args ...string) (string, error) {
		if len(args) != 1 {
			return "", fmt.Errorf("pathEscape requires 1 argument, but got (%d) arguments instead", len(args))
		}

		return url.PathEscape(args[0]), nil
	})

	return f.Eval(ipol.Eval(text))
}

func (o PublishTweetOpts) validReplyTo() bool {
	tweetID := o.ReplyTo

	parsedURL, err := url.Parse(o.ReplyTo)
	if err == nil {
		parts := strings.Split(parsedURL.Path, "/")
		if len(parts) < 1 {
			return false
		}
		tweetID = parts[len(parts)-1]
	}

	return len(tweetID) == 19 && allCharsNumeric(tweetID)
}

func (o PublishTweetOpts) validUrl() bool {
	_, err := url.Parse(o.Url)
	return err == nil
}

func (o PublishTweetOpts) JsonFmts() []string {
	partsA := strings.Split(o.Text, "{{")
	if len(partsA) == 0 {
		return []string{}
	}

	var jsonFmts []string

	for _, a := range partsA[1:] {
		partsB := strings.Split(a, "}}")
		if len(partsB) == 0 {
			continue
		}

		trimmed := strings.TrimSpace(partsB[0])
		jsonFmts = append(jsonFmts, trimmed)
	}

	return jsonFmts
}

type TwitterClient struct {
	*gotwi.Client
}

func newTwitterClient(creds TwitterAPICreds) (*TwitterClient, error) {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		APIKey:               creds.APIKey,
		APIKeySecret:         creds.APIKeySecret,
		OAuthToken:           creds.OAuthToken,
		OAuthTokenSecret:     creds.OAuthTokenSecret,
	}
	client, err := gotwi.NewClient(in)
	if err != nil {
		return nil, err
	}

	return &TwitterClient{
		Client: client,
	}, nil
}

func (c *TwitterClient) publishTweet(text string) (*managetweetTypes.CreateOutput, error) {
	p := &managetweetTypes.CreateInput{
		Text: gotwi.String(text),
	}

	return managetweet.Create(context.Background(), c.Client, p)
}

func (c *TwitterClient) publishTweetReply(text, tweetID string) (*managetweetTypes.CreateOutput, error) {
	p := &managetweetTypes.CreateInput{
		Text: gotwi.String(text),
		Reply: &managetweetTypes.CreateInputReply{
			InReplyToTweetID: tweetID,
		},
	}

	return managetweet.Create(context.Background(), c.Client, p)
}

func (c *TwitterClient) handle(opts PublishTweetOpts) (*managetweetTypes.CreateOutput, error) {
	var text = ""
	switch opts.PublishTweetType {
	case PublishTweetTypeText:
		text = opts.Text
	case PublishTweetTypeFetchJson:
		if !opts.validUrl() {
			return nil, fmt.Errorf("invalid url: %s", opts.Url)
		}

		resp, err := http.Get(opts.Url)
		if err != nil {
			return nil, err
		}

		text, err = opts.handleFetchJsonResp(resp)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	if text == "" {
		return nil, errors.New("tweet text cannot be an empty string")
	}

	if opts.validReplyTo() {
		return c.publishTweetReply(text, opts.ReplyTo)
	}

	return c.publishTweet(text)
}
