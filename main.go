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

type vehicle struct {
	id    int
	model string
	kind  string
}

type vehicleJSON map[string]interface{}

var config configuration
var db *sql.DB

func readConfig() {

	const CONFING = "conf.json"

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

}

func init() {

	readConfig()

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

	// Set headers to json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	rows, err := db.Query(queryVehicles)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	//vehicles := make([]*vehicle, 0)
	vehicles := make([]vehicle, 0)
	vehiclesJSON := make([]vehicleJSON, 0)
	for rows.Next() {
		//veh := new(vehicle)
		var veh vehicle
		var id, model, kind string
		err := rows.Scan(&id, &model, &kind)
		vehJSON := make(vehicleJSON)
		vehJSON["id"] = id
		vehJSON["type"] = kind
		vehJSON["model"] = model
		//err := rows.Scan(&veh.id, &veh.model, &veh.kind)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		vehicles = append(vehicles, veh)
		vehiclesJSON = append(vehiclesJSON, vehJSON)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	//b, err := json.Marshal(vehicles)
	//fmt.Fprint(w, b)
	// String(b)

	/*
		vvv := make(map[string]interface{})
		vvv["id"] = "1"
		vvv["type"] = "car"
		vvv["model"] = "nissan"
	*/
	/*
		for _, veh := range vehicles {
			//!!! fmt.Fprintf(w, "%d. %s (%s)\n", veh.id, veh.model, veh.kind)
			fmt.Printf("%d. %s (%s)\n", veh.id, veh.model, veh.kind)
			//json.NewEncoder(w).Encode(veh)
		}


		type fruits map[string]int
		fr := make(map[string]int)
		fr["apples"] = 25
		fr["oranges"] = 10
	*/

	println("marshaling...")
	b, err := json.Marshal(vehiclesJSON)
	//err = json.NewEncoder(w).Encode(&t)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	os.Stdout.Write(b)
	w.Write(b)

	//fmt.Fprintf(w, `{"Name":"Alice","Body":"Hello","Time":1294706395881547000}`)

}
