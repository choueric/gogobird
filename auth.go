package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
)

type UserInfo struct {
	Name   string
	Token  string
	Secret string
}

const configFile string = "./config.json"

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
	accessToken, vals, err := anaconda.GetCredentials(&requestToken, pin)
	if err != nil {
		fmt.Printf("GetCredentials failed: %v\n", err)
		return "", false
	}

	names := vals["screen_name"]
	if len(names) != 0 {
		name = names[0]
	}
	storeCredential(accessToken, name)
	return name, true
}

func readCredential() (UserInfo, error) {
	info := UserInfo{}
	file, err := os.Open(configFile)
	defer file.Close()
	if err != nil {
		fmt.Printf("open error: %v\n", err)
		return info, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&info)
	if err != nil {
		fmt.Println("json decode error:", err)
		return info, err
	}
	return info, nil
}

func storeCredential(credential *oauth.Credentials, name string) bool {
	info := UserInfo{
		Name:   name,
		Token:  credential.Token,
		Secret: credential.Secret,
	}

	b, err := json.Marshal(info)
	if err != nil {
		fmt.Println("Marshal error:", err)
		return false
	}

	file, err := os.Create(configFile)
	if err != nil {
		fmt.Printf("open error: %v\n", err)
		return false
	}
	defer file.Close()

	file.Write(b)
	return true
}

func ReadAccessCredential() (UserInfo, error) {
	return readCredential()
}
