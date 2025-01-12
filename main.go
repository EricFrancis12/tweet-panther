package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if err := SafeLoadEnvs(".env", ".env.local"); err != nil {
		log.Fatalf("error loading .env files: %s", err.Error())
	}

	var (
		Port             string = os.Getenv(EnvPort)
		AuthToken        string = os.Getenv(EnvAuthToken)
		APIKey           string = os.Getenv(EnvAPIKey)
		APIKeySecret     string = os.Getenv(EnvAPIKeySecret)
		OAuthToken       string = os.Getenv(EnvOAuthToken)
		OAuthTokenSecret string = os.Getenv(EnvOAuthTokenSecret)
	)

	if !isValidAuthToken(AuthToken) {
		log.Fatalf("invalid or missing variable (%s) from .env", EnvAuthToken)
	}

	creds := TwitterAPICreds{
		APIKey:           APIKey,
		APIKeySecret:     APIKeySecret,
		OAuthToken:       OAuthToken,
		OAuthTokenSecret: OAuthTokenSecret,
	}
	if !creds.isValid() {
		log.Fatalf(
			"required Twitter API credentials from .env: (%s), (%s), (%s), (%s)",
			EnvAPIKey, EnvAPIKeySecret, EnvOAuthToken, EnvOAuthTokenSecret,
		)
	}

	api, err := newAPI(Port, AuthToken, creds)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("API running at %s\n", Port)
	if err := api.run(); err != nil {
		log.Fatal(err)
	}
}
