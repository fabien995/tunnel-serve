package main

import (
	"github.com/fabien995/tunnel-serve/client"
	"os"
	"io/ioutil"
	"encoding/json"
)

type ClientConfig struct {
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
	client.Run(config.ReverseTunnelAddr, config.LocalServiceAddr)
}
