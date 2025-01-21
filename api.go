package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	handler    http.Handler
	router     *mux.Router
	client     *TwitterClient
	*Logger
}

func newAPI(listenAddr string, authToken string, creds ...TwitterAPICreds) (*API, error) {
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
		Logger:     newLogger(),
	}
	api.handler = api
	api.init()

	return api, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.Infof("%s @ %s (%s)\n", r.Method, r.URL.Path, r.RemoteAddr)
	a.router.ServeHTTP(w, r)
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

	a.router.HandleFunc("/api/users/by/username/{targetUsername}", a.auth(a.handleGetUserByUsername)).Methods(http.MethodGet)
	a.router.HandleFunc("/api/users/{targetUserID}", a.auth(a.handleGetUserByID)).Methods(http.MethodGet)
	a.router.HandleFunc("/api/users/{targetUserID}/tweets", a.auth(a.handleGetUserTweets)).Methods(http.MethodGet)

	a.router.HandleFunc("/healthz", a.handleHealthz)
	for _, path := range []string{"/", `/{catchAll:[a-zA-Z0-9=\-\/.]+}`} {
		a.router.HandleFunc(path, a.handleCatchAll)
	}
}

func (a *API) run() error {
	return http.ListenAndServe(a.listenAddr, a)
}

func (a *API) handlePublishTweet(w http.ResponseWriter, r *http.Request) {
	var opts PublishTweetOpts
	err := json.NewDecoder(r.Body).Decode(&opts)
	if err != nil {
		a.Errorf("error decoding json: %s\n", err.Error())
		writeBadRequest(w, nil)
		return
	}

	a.Infoln(opts.String())

	output, err := a.client.publishTweet(opts)
	if err != nil {
		a.LogErr(err)
		writeInternalServerError(w, nil)
		return
	}

	a.Infof("Published new Tweet (%s): %s\n", *output.Data.ID, *output.Data.Text)
	writeOK(w, output)
}

func (a *API) handleGetUserByUsername(w http.ResponseWriter, r *http.Request) {
	targetUsername := mux.Vars(r)[MuxVarTargetUsername]
	if targetUsername == "" {
		a.Errorf("missing path variable (%s)\n", MuxVarTargetUsername)
		writeBadRequest(w, nil)
		return
	}

	username := r.URL.Query().Get(QueryParamUsername)
	output, err := a.client.getUserByUsername(username, targetUsername)
	if err != nil {
		a.LogErr(err)
		writeInternalServerError(w, nil)
		return
	}

	a.Infof("Retrieved user by username (%s)\n", targetUsername)
	writeOK(w, output)
}

func (a *API) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	targetUserID := mux.Vars(r)[MuxVarTargetUserID]
	if targetUserID == "" {
		a.Errorf("missing path variable (%s)\n", MuxVarTargetUserID)
		writeBadRequest(w, nil)
		return
	}

	username := r.URL.Query().Get(QueryParamUsername)
	output, err := a.client.getUserByID(username, targetUserID)
	if err != nil {
		a.LogErr(err)
		writeInternalServerError(w, nil)
		return
	}

	a.Infof("Retrieved user by user ID (%s)\n", targetUserID)
	writeOK(w, output)
}

func (a *API) handleGetUserTweets(w http.ResponseWriter, r *http.Request) {
	targetUserID := mux.Vars(r)[MuxVarTargetUserID]
	if targetUserID == "" {
		a.Errorf("missing path variable (%s)\n", MuxVarTargetUserID)
		writeBadRequest(w, nil)
		return
	}

	username := r.URL.Query().Get(QueryParamUsername)
	output, err := a.client.getUserSingleTweets(username, targetUserID)
	if err != nil {
		a.LogErr(err)
		writeInternalServerError(w, nil)
		return
	}

	a.Infof("Retrieved (%d) tweets from user (%s)\n", len(output.Includes.Tweets), username)
	writeOK(w, output)
}

func (a *API) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, struct{}{})
}

func (a *API) handleCatchAll(w http.ResponseWriter, r *http.Request) {
	redirectUrl := os.Getenv(EnvCatchAllRedirectUrl)
	if isValidUrl(redirectUrl) {
		a.Infof("Redirecting visitor to %s\n", redirectUrl)
		redirectVisitor(w, r, redirectUrl)
		return
	}

	a.Infof("Route not found: %s\n", r.URL.Path)
	writeNotFound(w)
}
