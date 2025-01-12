package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type APIResp struct {
	Success bool   `json:"success,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func newAPIResp(success bool, msg string, data any) *APIResp {
	return &APIResp{
		Success: success,
		Msg:     msg,
		Data:    data,
	}
}

type API struct {
	listenAddr string
	authToken  string
	router     *mux.Router
	client     *TwitterClient
}

func newAPI(listenAddr string, authToken string, creds TwitterAPICreds) (*API, error) {
	la := ensurePrefix(listenAddr, ":")
	if !allCharsNumeric(la[1:]) {
		return nil, fmt.Errorf("invalid listen address: %s", listenAddr)
	}

	client, err := newTwitterClient(creds)
	if err != nil {
		return nil, err
	}

	api := &API{
		listenAddr: la,
		authToken:  authToken,
		router:     mux.NewRouter(),
		client:     client,
	}
	api.init()

	return api, nil
}

func (a *API) auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		value := r.Header.Get(HTTPHeaderAuthorization)
		parts := strings.Split(value, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			authToken := parts[1]
			if authToken == a.authToken {
				h(w, r)
				return
			}
		}

		writeUnauthorized(w, nil)
	}
}

func (a *API) init() {
	a.router.HandleFunc("/api/tweet", a.auth(a.handlePublishTweet)).Methods(http.MethodPost)
}

func (a *API) run() error {
	return http.ListenAndServe(a.listenAddr, a.router)
}

func (a *API) handlePublishTweet(w http.ResponseWriter, r *http.Request) {
	var opts PublishTweetOpts
	err := json.NewDecoder(r.Body).Decode(&opts)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	output, err := a.client.handle(opts)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeOK(w, output)
}
