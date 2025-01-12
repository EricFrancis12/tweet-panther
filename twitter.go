package main

import (
	"context"
	"net/url"
	"strings"

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

func (tac TwitterAPICreds) areValid() bool {
	return tac.APIKey != "" && tac.APIKeySecret != "" && tac.OAuthToken != "" && tac.OAuthTokenSecret != ""
}

type PublishTweetOpts struct {
	Text    string `json:"text"`
	ReplyTo string `json:"replyTo"`
}

func (o *PublishTweetOpts) isValidReplyTo() bool {
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
	if opts.isValidReplyTo() {
		return c.publishTweetReply(opts.Text, opts.ReplyTo)
	}

	return c.publishTweet(opts.Text)
}
