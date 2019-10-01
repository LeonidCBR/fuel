package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: implement work with SQLite3

// TODO: implement REST API using SSL

// TODO: implement Auth /api/user/new using JWT

// using like https://localhost:8585/api/v1/fuel/vehicles/1

// example of mux https://metanit.com/go/web/1.4.php

type configuration struct {
	IP          string
	Port        int
	UseTLS      bool
	Certificate string
	Key         string
	DB          string
}

type vehicle struct {
	id    int
	model string
	kind  string
}

func main() {

	const CONFING = "conf.json"

	println("Reading config file (conf.json)")

	file, err := os.Open(CONFING)
	if err != nil {
		log.Fatalln("Can not open config file", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	config := configuration{}

	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalln("Error while decoding:", err)
	}

	db, err := sql.Open("sqlite3", "fuel.db")
	if err != nil {
		log.Fatalln("Can not connect to DB:", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, model, type FROM vehicles")
	if err != nil {
		log.Fatalln("Error while selecting:", err)
	}
	defer rows.Close()

	vehicles := make([]*vehicle, 0)
	for rows.Next() {
		veh := new(vehicle)
		err := rows.Scan(&veh.id, &veh.model, &veh.kind)
		if err != nil {
			log.Fatalln("Scan error:", err)
		}
		vehicles = append(vehicles, veh)
	}
	if err = rows.Err(); err != nil {
		log.Fatalln("Rows error:", err)
	}

	for _, veh := range vehicles {
		fmt.Printf("%d. %s (%s)\n", veh.id, veh.model, veh.kind)
	}
	/*
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
	*/
}
