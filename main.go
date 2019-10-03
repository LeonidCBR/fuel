package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: implement Auth /api/user/new using JWT

// using like:
// https://localhost:8585/api/v1/fuel/vehicles
// https://localhost:8585/api/v1/fuel/vehicles/create
// https://localhost:8585/api/v1/fuel/vehicles/show?id=3

// example of mux https://metanit.com/go/web/1.4.php

type configuration struct {
	IP          string
	Port        int
	UseTLS      bool
	Certificate string
	Key         string
	DB          string
}

type Vehicle struct {
	ID    int    `json:"id"`
	Model string `json:"model"`
	Kind  string `json:"type"`
}

type Response struct {
	Status   string    `json:"status"`
	Total    int       `json:"total"`
	Vehicles []Vehicle `json:"vehicles"`
}

var config configuration

var db *sql.DB

func readConfig() {

	const CONFING = "fuel.conf"

	fmt.Printf("Reading config file (%s)\n", CONFING)

	file, err := os.Open(CONFING)
	if err != nil {
		log.Fatalln("Can not open config file", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalln("Error while decoding:", err)
	}

	// TODO: check files existing in folders from config

}

func init() {

	readConfig()

	fmt.Println("DB file:", config.DB)

	var err error
	db, err = sql.Open("sqlite3", config.DB)
	if err != nil {
		log.Fatalln("Can not connect to DB:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalln(err)
	}
}

func main() {

	http.HandleFunc("/api/v1/fuel/vehicles", vehiclesIndex)

	listenAddress := fmt.Sprintf("%s:%d", config.IP, config.Port)

	if config.UseTLS {
		fmt.Printf("Starting service at https://%s\n", listenAddress)
		http.ListenAndServeTLS(listenAddress, config.Certificate, config.Key, nil)
	} else {
		fmt.Printf("Starting service at http://%s\n", listenAddress)
		http.ListenAndServe(listenAddress, nil)
	}

}

func vehiclesIndex(w http.ResponseWriter, r *http.Request) {

	const queryVehicles = "SELECT id, model, type FROM vehicles"

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	// Set headers to json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	rows, err := db.Query(queryVehicles)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	response := new(Response)
	total := 0
	for rows.Next() {
		var vehicle Vehicle
		err := rows.Scan(&vehicle.ID, &vehicle.Model, &vehicle.Kind)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		response.Vehicles = append(response.Vehicles, vehicle)
		total++
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	/*
		result, err := json.Marshal(response)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		w.Write(result)
	*/

	response.Status = "ok"
	response.Total = total

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

}
