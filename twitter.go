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
	"github.com/michimani/gotwi/tweet/timeline"
	timelineTypes "github.com/michimani/gotwi/tweet/timeline/types"
)

type TwitterAPICreds struct {
	UserID           string
	APIKey           string
	APIKeySecret     string
	OAuthToken       string
	OAuthTokenSecret string
}

func (tac TwitterAPICreds) isValid() bool {
	return tac.UserID != "" && tac.APIKey != "" && tac.APIKeySecret != "" && tac.OAuthToken != "" && tac.OAuthTokenSecret != ""
}

func (tac TwitterAPICreds) String() string {
	return fmt.Sprintf(
		"TwitterAPICreds{ APIKey: %s, APIKeySecret: %s, OAuthToken: %s, OAuthTokenSecret: %s }",
		tac.APIKey,
		tac.APIKeySecret,
		tac.OAuthToken,
		tac.OAuthTokenSecret,
	)
}

type PublishTweetOpts struct {
	PublishTweetType PublishTweetType `json:"publishTweetType"`
	Text             string           `json:"text"`
	ReplyTo          string           `json:"replyTo"`
	Url              string           `json:"url"`
	UserID           string           `json:"userId"`
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
		level = data
		text  = o.Text
		ipol  = stripol.New("{*{", "}*}")
	)

	for _, jsonFmt := range o.JsonFmts() {
		keys := strings.Split(jsonFmt, ".")
		for _, key := range keys {
			m, ok := level.(map[string]interface{})
			if !ok {
				return "", errors.New("invalid jsonFmt A")
			}

			if level, ok = m[key]; !ok {
				return "", errors.New("invalid jsonFmt B")
			}
		}

		var s string

		switch d := level.(type) {
		case string:
			s = d
		case int64:
			s = fmt.Sprintf("%d", d)
		case float64:
			s = fmt.Sprintf("%f", d)
		case bool:
			s = strconv.FormatBool(d)
		default:
			b, err := json.Marshal(level)
			if err != nil {
				return "", err
			}
			s = string(b)
		}

		ipol.RegisterVar(jsonFmt, s)
		level = data
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

func (o PublishTweetOpts) getReplyToTweetID() (string, error) {
	if o.ReplyTo == "" {
		return "", errors.New("replyTo is an empty string")
	}

	tweetID := o.ReplyTo

	parsedURL, err := url.Parse(o.ReplyTo)
	if err == nil {
		path := remSuffixIfExists(parsedURL.Path, "/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 {
			return "", fmt.Errorf("expected length of at least 1, but got: %d", len(parts))
		}
		tweetID = parts[len(parts)-1]
	}

	if len(tweetID) == 19 && allCharsNumeric(tweetID) {
		return tweetID, nil
	}

	return "", fmt.Errorf("tweet ID must be 19 characters long, and contain only numeric characters (received: %s)", tweetID)
}

func (o PublishTweetOpts) validReplyTo() bool {
	_, err := o.getReplyToTweetID()
	return err == nil
}

func (o PublishTweetOpts) validUrl() bool {
	return isValidUrl(o.Url)
}

func (o PublishTweetOpts) JsonFmts() []string {
	partsA := strings.Split(o.Text, "{*{")
	if len(partsA) == 0 {
		return []string{}
	}

	var jsonFmts []string

	for _, a := range partsA[1:] {
		partsB := strings.Split(a, "}*}")
		if len(partsB) == 0 {
			continue
		}

		trimmed := strings.TrimSpace(partsB[0])
		jsonFmts = append(jsonFmts, trimmed)
	}

	return jsonFmts
}

func (o PublishTweetOpts) String() string {
	return fmt.Sprintf(
		"PublishTweetOpts{ PublishTweetType: %s, Text: %s, ReplyTo: %s, Url: %s }",
		o.PublishTweetType,
		o.Text,
		o.ReplyTo,
		o.Url,
	)
}

type TwitterClient struct {
	clients map[TwitterAPICreds]*gotwi.Client
}

func newTwitterClient(creds []TwitterAPICreds) (*TwitterClient, error) {
	if len(creds) == 0 {
		return nil, errors.New("at least (1) Twitter API Cred is required")
	}

	var clients = make(map[TwitterAPICreds]*gotwi.Client)
	for _, cred := range creds {
		if !cred.isValid() {
			return nil, fmt.Errorf("invalid Twitter API cred: %s", cred.String())
		}

		in := &gotwi.NewClientInput{
			AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
			APIKey:               cred.APIKey,
			APIKeySecret:         cred.APIKeySecret,
			OAuthToken:           cred.OAuthToken,
			OAuthTokenSecret:     cred.OAuthTokenSecret,
		}
		client, err := gotwi.NewClient(in)
		if err != nil {
			return nil, err
		}

		clients[cred] = client
	}

	return &TwitterClient{
		clients: clients,
	}, nil
}

func (c *TwitterClient) getClientByUserID(userID string) (*gotwi.Client, bool) {
	if userID == "" {
		return nil, false
	}

	for cred, client := range c.clients {
		if cred.UserID == userID {
			return client, true
		}
	}

	return nil, false
}

func (c *TwitterClient) doCreate(userID string, p *managetweetTypes.CreateInput) (*managetweetTypes.CreateOutput, error) {
	if userID != "" {
		if client, ok := c.getClientByUserID(userID); ok {
			return managetweet.Create(context.Background(), client, p)
		}
		return nil, fmt.Errorf("userID (%s) not found in client pool", userID)
	}

	for _, client := range c.clients {
		output, err := managetweet.Create(context.Background(), client, p)
		if err == nil {
			return output, nil
		} else if !isRateLimitErr(err) {
			return nil, fmt.Errorf("error publishing tweet ( %s ): %s", *p.Text, err.Error())
		}
	}

	return nil, fmt.Errorf(
		"error creating tweet ( %s ): all %d Twitter clients were rate-limited",
		*p.Text,
		len(c.clients),
	)
}

func (c *TwitterClient) publishNewTweet(userID, text string) (*managetweetTypes.CreateOutput, error) {
	p := &managetweetTypes.CreateInput{
		Text: gotwi.String(text),
	}
	return c.doCreate(userID, p)
}

func (c *TwitterClient) publishTweetReply(userID, text, tweetID string) (*managetweetTypes.CreateOutput, error) {
	p := &managetweetTypes.CreateInput{
		Text: gotwi.String(text),
		Reply: &managetweetTypes.CreateInputReply{
			InReplyToTweetID: tweetID,
		},
	}
	return c.doCreate(userID, p)
}

func (c *TwitterClient) publishTweet(opts PublishTweetOpts) (*managetweetTypes.CreateOutput, error) {
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
		tweetID, err := opts.getReplyToTweetID()
		if err != nil {
			return nil, err
		}
		return c.publishTweetReply(opts.UserID, text, tweetID)
	}

	return c.publishNewTweet(opts.UserID, text)
}

func (c *TwitterClient) getUserTweets(userID, targetUserID string) (*timelineTypes.ListTweetsOutput, error) {
	if targetUserID == "" {
		return nil, errors.New("missing targetUserID")
	}

	p := &timelineTypes.ListTweetsInput{
		ID: targetUserID,
	}

	if userID != "" {
		if client, ok := c.getClientByUserID(userID); ok {
			// TODO: fix timeline.ListTweets() returning 400 lever error (invalid request)
			return timeline.ListTweets(context.Background(), client, p)
		}
		return nil, fmt.Errorf("userID (%s) not found in client pool", userID)
	}

	for _, client := range c.clients {
		output, err := timeline.ListTweets(context.Background(), client, p)
		if err == nil {
			return output, nil
		} else if !isRateLimitErr(err) {
			return nil, fmt.Errorf("error getting tweets for user ( %s ): %s", targetUserID, err.Error())
		}
	}

	return nil, fmt.Errorf(
		"error getting tweets for user ( %s ): all (%d) Twitter client(s) were rate-limited",
		targetUserID,
		len(c.clients),
	)
}
