package main

import (
	"fmt"

	"github.com/choueric/jconfig"
)

const DefContent = `{
	"proxy":"127.0.0.1:1080",
	"user":"",
	"consumer_key": "",
	"consumer_secret": "",
	"access_token":"",
	"access_token_secret":""
}
`

type Config struct {
	ProxyAddr         string `json:"proxy"`
	UserName          string `json:"user"`
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	AccessToken       string `json:"access_token""`
	AccessTokenSecret string `json:"access_token_secret""`
}

func getConfig() *Config {
	jc := jconfig.New("config.json", Config{})

	if _, err := jc.Load(DefContent); err != nil {
		fmt.Println("load config error:", err)
		return nil
	}

	return jc.Data().(*Config)
}
