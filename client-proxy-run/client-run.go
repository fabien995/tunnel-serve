package main

import (
	"github.com/fabien995/tunnel-serve/clientProxy"
	"os"
	"io/ioutil"
	"encoding/json"
)

type ClientConfig struct {
	ProxyPort string
	ReverseTunnelAddr string
	LocalServiceAddr string
}

func main() {
	// Parse JSON.
	jsonFile, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	var config ClientConfig
	defer jsonFile.Close()
	jsonFileBytes, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(jsonFileBytes, &config)
	clientProxy.Run(config.ProxyPort, config.ReverseTunnelAddr, config.LocalServiceAddr)
}
