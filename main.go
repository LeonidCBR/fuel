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

// TODO: get config name from args

// TODO: check files existing in folders from config

// TODO: implement Auth /api/user/new using JWT

/*
	TODO: implement error returns
	HTTP/1.1 200 OK
	{
		"error": {
			"error_code": 5,
			"error_msg": "some error"
		}
	}

	vk example
	{"error":{"error_code":5,"error_msg":"User authorization failed: no access_token passed.","request_params":[{"key":"method","value":"getProfiles"},{"key":"oauth","value":"1"},{"key":"model","value":"camry"},{"key":"type","value":"bluecar"}]}}
*/

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
	ID    int64  `json:"id"`
	Model string `json:"model"`
	Kind  string `json:"type"`
}

type GasUp struct {
	ID            int64   `json:"id"`
	VehicleID     int64   `json:"id_vehicle"`
	RefuelingDate string  `json:"refueling_date"`
	Liters        float32 `json:"liters"`
	Cost          float32 `json:"cost"`
	Odometer      int     `json:"odometer"`
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

	// TODO: get config name from args

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

	// get list of vehicles
	http.HandleFunc("/api/v1/fuel/vehicles", vehiclesIndex)

	// add new vehicle
	http.HandleFunc("/api/v1/fuel/vehicles/create", vehiclesCreate)

	// add new gas up
	http.HandleFunc("/api/v1/fuel/gasup/create", gasUpCreate)

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

	response.Status = "ok"
	response.Total = total

	result, err := json.Marshal(response)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)

	/*
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
	*/
}

func vehiclesCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(415), 415)
		return
	}

	/*
		// when Content-Type is application/x-www-form-urlencoded
		model := r.FormValue("model")
		kind := r.FormValue("type")
		fmt.Printf("model: %s\ntype: %s\n", model, kind)
	*/

	// unmarshal json data from body
	var vehicle Vehicle
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&vehicle)
	if err != nil {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	if vehicle.Model == "" || vehicle.Kind == "" {
		//http.Error(w, http.StatusText(400), 400)
		sendError(w, 1)
		return
	}

	//fmt.Printf("%s (%s)\n", vehicle.Model, vehicle.Kind)
	//price, err := strconv.ParseFloat("80.25", 32)

	result, err := db.Exec("INSERT INTO vehicles (model, type) VALUES($1, $2)", vehicle.Model, vehicle.Kind)
	if err != nil {
		//http.Error(w, http.StatusText(500), 500)
		sendError(w, 2)
		return
	}

	newID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	vehicle.ID = newID
	//fmt.Printf("ok! id:%d\n", newID)

	// make response to client

	res, err := json.Marshal(vehicle)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(res)

}

func gasUpCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(415), 415)
		return
	}

	// unmarshal json data from body
	var gasUp GasUp
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&gasUp)
	if err != nil {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	//fmt.Printf("(%d, %s, %.2f, %.2f, %d)", gasUp.VehicleID, gasUp.RefuelingDate, gasUp.Liters, gasUp.Cost, gasUp.Odometer)

	if gasUp.VehicleID == 0 || gasUp.RefuelingDate == "" || gasUp.Liters == 0 || gasUp.Cost == 0 || gasUp.Odometer == 0 {
		sendError(w, 1)
		return
	}

	result, err := db.Exec("INSERT INTO fuel (id_vehicle, refueling_date, liters, cost, odometer) VALUES($1, $2, $3, $4, $5)", gasUp.VehicleID, gasUp.RefuelingDate, gasUp.Liters, gasUp.Cost, gasUp.Odometer)
	if err != nil {
		sendError(w, 2)
		return
	}

	newID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	gasUp.ID = newID

	// make response to client

	res, err := json.Marshal(gasUp)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(res)

}

func sendError(w http.ResponseWriter, errCode int) {

	type ResponseError struct {
		Code    int    `json:"error_code"`
		Message string `json:"error_msg"`
	}

	response := new(ResponseError)
	response.Code = errCode

	switch errCode {
	case 1:
		response.Message = "Bad request params"
	case 2:
		response.Message = "Can not insert data into DB"
	default:
		response.Message = "Unknown error"
	}

	result, err := json.Marshal(response)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)

	/*
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
	*/
}
