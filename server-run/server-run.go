package main

import (
	"github.com/fabien995/tunnel-serve/server"
	"strconv"
	"os"
	"io/ioutil"
	"encoding/json"
)

type ServerConfig struct {
	BindAddress string
	ControlPort string
}

func main() {
	// Parse JSON.
	jsonFile, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	var config ServerConfig
	defer jsonFile.Close()
	jsonFileBytes, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(jsonFileBytes, &config)
	bindAddress := config.BindAddress
	controlPort, err := strconv.Atoi(config.ControlPort)
	if err != nil {
		panic(err)
	}
	server.Run(bindAddress, controlPort)
}
