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
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/tweet/managetweet"
	managetweetTypes "github.com/michimani/gotwi/tweet/managetweet/types"
	"github.com/michimani/gotwi/tweet/timeline"
	timelineTypes "github.com/michimani/gotwi/tweet/timeline/types"
	"github.com/michimani/gotwi/tweet/tweetlookup"
	tweetlookupTypes "github.com/michimani/gotwi/tweet/tweetlookup/types"
	"github.com/michimani/gotwi/user/userlookup"
	userlookupTypes "github.com/michimani/gotwi/user/userlookup/types"
)

type TwitterAPICreds struct {
	Username         string
	APIKey           string
	APIKeySecret     string
	OAuthToken       string
	OAuthTokenSecret string
}

func (tac TwitterAPICreds) isValid() bool {
	return tac.Username != "" && tac.APIKey != "" && tac.APIKeySecret != "" && tac.OAuthToken != "" && tac.OAuthTokenSecret != ""
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
	Username         string           `json:"username"`
	IgnoreReplies    bool             `json:"ignoreReplies"`
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

func (c *TwitterClient) getClientByUsername(username string) (*gotwi.Client, bool) {
	if username == "" {
		return nil, false
	}

	for cred, client := range c.clients {
		if cred.Username == username {
			return client, true
		}
	}

	return nil, false
}

func (c *TwitterClient) tweetIsReply(tweetID string) (bool, error) {
	if !isValidTweetID(tweetID) {
		return false, fmt.Errorf("invalid tweetID: (%s)", tweetID)
	}

	p := &tweetlookupTypes.GetInput{
		ID: tweetID,
	}

	for _, client := range c.clients {
		output, err := tweetlookup.Get(context.Background(), client, p)
		if err == nil {
			return output.Data.InReplyToUserID != nil && *output.Data.InReplyToUserID != "", nil
		} else if !isRateLimitErr(err) {
			return false, fmt.Errorf("error checking if tweet ( %s ) is a reply: %s", p.ID, err.Error())
		}
	}

	return false, fmt.Errorf(
		"error checking if tweet ( %s ) is a reply: all %d Twitter clients were rate-limited",
		p.ID,
		len(c.clients),
	)
}

func (c *TwitterClient) doCreate(username string, p *managetweetTypes.CreateInput) (*managetweetTypes.CreateOutput, error) {
	if username != "" {
		if client, ok := c.getClientByUsername(username); ok {
			return managetweet.Create(context.Background(), client, p)
		}
		return nil, fmt.Errorf("username (%s) not found in client pool", username)
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

func (c *TwitterClient) publishTweetSingle(username, text string) (*managetweetTypes.CreateOutput, error) {
	p := &managetweetTypes.CreateInput{
		Text: gotwi.String(text),
	}
	return c.doCreate(username, p)
}

func (c *TwitterClient) publishTweetReply(username, text, tweetID string) (*managetweetTypes.CreateOutput, error) {
	p := &managetweetTypes.CreateInput{
		Text: gotwi.String(text),
		Reply: &managetweetTypes.CreateInputReply{
			InReplyToTweetID: tweetID,
		},
	}
	return c.doCreate(username, p)
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

		if !isValidTweetID(tweetID) {
			return nil, fmt.Errorf("invalid tweetID: (%s)", tweetID)
		}

		if opts.IgnoreReplies {
			isReply, err := c.tweetIsReply(tweetID)
			if err != nil {
				return nil, err
			}

			if isReply {
				return nil, fmt.Errorf("tweet ID (%s) is a reply", tweetID)
			}
		}

		return c.publishTweetReply(opts.Username, text, tweetID)
	}

	return c.publishTweetSingle(opts.Username, text)
}

func (c *TwitterClient) getUserByUsername(username, targetUsername string) (*userlookupTypes.GetByUsernameOutput, error) {
	if targetUsername == "" {
		return nil, errors.New("missing targetUsername")
	}

	p := &userlookupTypes.GetByUsernameInput{
		Username: targetUsername,
	}

	if username != "" {
		if client, ok := c.getClientByUsername(username); ok {
			return userlookup.GetByUsername(context.Background(), client, p)
		}
		return nil, fmt.Errorf("username (%s) not found in client pool", username)
	}

	for _, client := range c.clients {
		output, err := userlookup.GetByUsername(context.Background(), client, p)
		if err == nil {
			return output, nil
		} else if !isRateLimitErr(err) {
			return nil, fmt.Errorf("error getting user by username ( %s ): %s", targetUsername, err.Error())
		}
	}

	return nil, fmt.Errorf(
		"error getting user by username ( %s ): all (%d) Twitter client(s) were rate-limited",
		targetUsername,
		len(c.clients),
	)
}

func (c *TwitterClient) getUserByID(username, targetUserID string) (*userlookupTypes.GetOutput, error) {
	if targetUserID == "" {
		return nil, errors.New("missing targetUserID")
	}

	p := &userlookupTypes.GetInput{
		ID: targetUserID,
	}

	if username != "" {
		if client, ok := c.getClientByUsername(username); ok {
			return userlookup.Get(context.Background(), client, p)
		}
		return nil, fmt.Errorf("username (%s) not found in client pool", username)
	}

	for _, client := range c.clients {
		output, err := userlookup.Get(context.Background(), client, p)
		if err == nil {
			return output, nil
		} else if !isRateLimitErr(err) {
			return nil, fmt.Errorf("error getting user by ID ( %s ): %s", targetUserID, err.Error())
		}
	}

	return nil, fmt.Errorf(
		"error getting user by ID ( %s ): all (%d) Twitter client(s) were rate-limited",
		targetUserID,
		len(c.clients),
	)
}

func (c *TwitterClient) getUserSingleTweets(username, targetUserID string) (*timelineTypes.ListTweetsOutput, error) {
	if targetUserID == "" {
		return nil, errors.New("missing targetUserID")
	}

	p := &timelineTypes.ListTweetsInput{
		ID: targetUserID,
		Exclude: fields.ExcludeList{
			fields.ExcludeReplies,
			fields.ExcludeRetweets,
		},
	}

	if username != "" {
		if client, ok := c.getClientByUsername(username); ok {
			return timeline.ListTweets(context.Background(), client, p)
		}
		return nil, fmt.Errorf("username (%s) not found in client pool", username)
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
