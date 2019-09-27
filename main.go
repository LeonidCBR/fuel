package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// TODO: implement work with SQLite3

// TODO: implement REST API using SSL

// TODO: implement Auth /api/user/new using JWT

// using like https://localhost:8585/api/v1/fuel/vehicles/1

// example of mux https://metanit.com/go/web/1.4.php

// Configuration of this program
type Configuration struct {
	IP          string
	Port        int
	UseTLS      bool
	Certificate string
	Key         string
}

func main() {

	const CONFING = "conf.json"

	println("Reading config file (conf.json)")

	file, err := os.Open(CONFING)
	if err != nil {
		println("Can not open config file")
		os.Exit(1)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	config := Configuration{}

	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error while decoding:", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	})

	listenAddress := fmt.Sprintf("%s:%d", config.IP, config.Port)
	fmt.Println("Starting service at", listenAddress)

	if !config.UseTLS {
		println("not use TLS")
		http.ListenAndServe(listenAddress, nil)
	} else {
		println("using TLS")
		http.ListenAndServeTLS(listenAddress, config.Certificate, config.Key, nil)
	}
}
