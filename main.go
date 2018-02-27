package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"net/http"
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
	
	indexMessage := "Hello from Hari"
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
	var testSystem System
	err := decoder.Decode(&testSystem)
	if err != nil {
		log.Printf("Error decoding register system POST: %v\n", err)
    	http.Error(w, "Error decoding POST body", http.StatusBadRequest)
		return
	}
	log.Printf("%+v", testSystem)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"id": 1}`))
}

func main() {
	port := os.Getenv("PORT")
	log.Printf("Retrieved Port: %v\n", port)

	http.HandleFunc("/", index)
	http.HandleFunc("/register/system", registerSystem)
	log.Printf("Listening for requests...\n")
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Fatal("$PORT must be set")
	}
}
