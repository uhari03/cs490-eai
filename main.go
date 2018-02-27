package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"net/http"

	_ "github.com/lib/pq"
)

var (
	db *sql.DB
)

type System struct {
	Name string `json:"systemName"`
	ApplicationEndpoint string `json:"applicationEndpoint"`
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
    	http.Error(w, "404 not found", http.StatusNotFound)
        return
    }
	
	indexMessage := "Hello from Hari\n"
	w.Write([]byte(indexMessage))
}

func registerSystem(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/register/system" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	} else if r.Method != "POST" {
		http.Error(w, "Only POST methods are supported", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var newSystemRow System
	err := decoder.Decode(&newSystemRow)
	if err != nil {
		log.Printf("Error decoding register system POST: %v\n", err)
    	http.Error(w, "Error decoding POST body", http.StatusBadRequest)
		return
	}
	log.Printf("%+v", newSystemRow)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS systems (name TEXT PRIMARY KEY NOT NULL, applicationEndpoint TEXT NOT NULL)")
	if err != nil {
		log.Printf("Error creating systems table: %v\n", err)
    	http.Error(w, "Error creating systems table", http.StatusInternalServerError)
		return
    }

	_, err = db.Exec(fmt.Sprintf("INSERT INTO systems VALUES ('%v', '%v')", newSystemRow.Name, newSystemRow.ApplicationEndpoint))
	if err != nil {
		log.Printf("Error creating new systems row: %v\n", err)
    	http.Error(w, "Error creating new systems row", http.StatusInternalServerError)
		return
    }

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"id\": 1}\n"))
}

func viewSystem(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/view/system" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	rows, err := db.Query("SELECT * FROM systems")
	if err != nil {
		log.Printf("Error querying systems table: %v\n", err)
    	http.Error(w, "Error querying systems table", http.StatusInternalServerError)
		return
    }
	defer rows.Close()

	var queriedSystemRows []System
	for rows.Next() {
        var queriedSystemRow System
        if err := rows.Scan(&(queriedSystemRow.Name), &(queriedSystemRow.ApplicationEndpoint)); err != nil {
			log.Printf("Error querying systems table: %v\n", err)
			http.Error(w, "Error querying systems table", http.StatusInternalServerError)
			return
        }
		queriedSystemRows = append(queriedSystemRows, queriedSystemRow)
    }

	b, err := json.Marshal(&queriedSystemRows)
	if err != nil {
		log.Printf("Error marshalling data: %v\n", err)
		http.Error(w, "Error marshalling data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func main() {
	port := os.Getenv("PORT")
	log.Printf("Retrieved Port: %v\n", port)

	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("Error opening database: %v\n", err)
    }

	http.HandleFunc("/", index)
	http.HandleFunc("/register/system", registerSystem)
	log.Printf("Listening for requests...\n")
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		log.Fatalf("Error starting HTTP server: %v\n", err)
	}
}
