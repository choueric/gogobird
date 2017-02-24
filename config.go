package main

import (
	"fmt"

	"github.com/choueric/jconfig"
)

const DefContent = `{
	"consumer_key": "",
	"consumer_secret": ""
}
`

type Config struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
}

func getConfig() *Config {
	jc := jconfig.New("config.json", Config{})

	if _, err := jc.Load(DefContent); err != nil {
		fmt.Println("load config error:", err)
		return nil
	}

	return jc.Data().(*Config)
}
