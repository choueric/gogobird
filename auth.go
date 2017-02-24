package main

import (
	"fmt"

	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
)

var (
	requestToken oauth.Credentials
)

func GetAuthUrl() (string, error) {
	url, token, err := anaconda.AuthorizationURL("oob")
	if err != nil {
		return "", err
	}

	if token == nil {
		fmt.Printf("Get Authorization URL is nil\n")
		return "", err
	}

	requestToken = *token
	return url, nil
}

func DoAuth(pin string) (name string, _ bool) {
	_, vals, err := anaconda.GetCredentials(&requestToken, pin)
	if err != nil {
		fmt.Printf("GetCredentials failed: %v\n", err)
		return "", false
	}

	names := vals["screen_name"]
	if len(names) != 0 {
		name = names[0]
	}
	return name, true
}
