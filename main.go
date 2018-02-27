package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"net/http"

	"github.com/lib/pq"
)

var (
	db *sql.DB
)

type System struct {
	Name string `json:"systemName"`
	ApplicationEndpoint string `json:"applicationEndpoint"`
}

type Topic struct {
	Name string `json:"topicName"`
	Description string `json:"description"`
	Owner string `json:"owner"`
	Structure string `json:"structure"`
	Subscribers []string `json:"subscribers"`
}

type Event struct {
	Topic string `json:"topicName"`
	Data string `json:"data"`
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

func registerTopic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/register/topic" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	} else if r.Method != "POST" {
		http.Error(w, "Only POST methods are supported", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var newTopicRow Topic
	err := decoder.Decode(&newTopicRow)
	if err != nil {
		log.Printf("Error decoding register topic POST: %v\n", err)
    	http.Error(w, "Error decoding POST body", http.StatusBadRequest)
		return
	}
	log.Printf("%+v", newTopicRow)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS topics (name TEXT PRIMARY KEY NOT NULL, description TEXT NOT NULL, owner TEXT NOT NULL, structure JSON NOT NULL, subscribers TEXT [])")
	if err != nil {
		log.Printf("Error creating topics table: %v\n", err)
    	http.Error(w, "Error creating topics table", http.StatusInternalServerError)
		return
    }

	_, err = db.Exec(fmt.Sprintf("INSERT INTO topics (name, description, owner, structure) VALUES ('%v', '%v', '%v', '%v')", newTopicRow.Name, newTopicRow.Description, newTopicRow.Owner, newTopicRow.Structure))
	if err != nil {
		log.Printf("Error creating new topics row: %v\n", err)
    	http.Error(w, "Error creating new topics row", http.StatusInternalServerError)
		return
    }

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"id\": 1}\n"))
}

func subscribe(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/subscribe" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	systemName := r.URL.Query().Get("systemName")
	if systemName == "" {
		log.Printf("Blank system name provided\n")
    	http.Error(w, "Must provide systemName parameter", http.StatusBadRequest)
		return
	}

	topicName := r.URL.Query().Get("topicName")
	if topicName == "" {
		log.Printf("Blank topic name provided\n")
    	http.Error(w, "Must provide topicName parameter", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(fmt.Sprintf("UPDATE topics SET subscribers = array_cat(subscribers, '{%v}') WHERE name = '%v'", systemName, topicName))
	if err != nil {
		log.Printf("Error updating topics row with new subscriber: %v\n", err)
    	http.Error(w, "Error updating topics row with new subscriber", http.StatusInternalServerError)
		return
    }

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Successfully subscribed\n"))
}

func publish(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/publish" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	} else if r.Method != "POST" {
		http.Error(w, "Only POST methods are supported", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var newEvent Event
	err := decoder.Decode(&newEvent)
	if err != nil {
		log.Printf("Error decoding register topic POST: %v\n", err)
    	http.Error(w, "Error decoding POST body", http.StatusBadRequest)
		return
	}
	log.Printf("%+v\n", newEvent)

	subscriptionsColumn, err := db.Query(fmt.Sprintf("SELECT subscribers FROM topics WHERE name = '%v'", newEvent.Topic))
	if err != nil {
		log.Printf("Error querying topics table: %v\n", err)
    	http.Error(w, "Error querying topics table", http.StatusInternalServerError)
		return
    }
	defer subscriptionsColumn.Close()

	var subscriptions []string
	for subscriptionsColumn.Next() {
        if err := subscriptionsColumn.Scan(pq.Array(&subscriptions)); err != nil {
			log.Printf("Error querying topics table: %v\n", err)
			http.Error(w, "Error querying topics table", http.StatusInternalServerError)
			return
        }
    }

	var subscriptionSearchString string
	for index, subscription := range subscriptions {
		if index == 0 {
			subscriptionSearchString = fmt.Sprintf("'%v'", subscription)
			continue
		}

		subscriptionSearchString += fmt.Sprintf(", '%v'", subscription)
	}
	log.Printf("%v\n", subscriptionSearchString)
	subscriberApplicationEndpointRows, err := db.Query(fmt.Sprintf("SELECT applicationEndpoint FROM systems WHERE name in (%v)", subscriptionSearchString))
	if err != nil {
		log.Printf("Error querying systems table: %v\n", err)
    	http.Error(w, "Error querying systems table", http.StatusInternalServerError)
		return
    }
	defer subscriberApplicationEndpointRows.Close()

	var output string
	for subscriberApplicationEndpointRows.Next() {
		var applicationEndpoint string
        if err := subscriberApplicationEndpointRows.Scan(&applicationEndpoint); err != nil {
			log.Printf("Error querying system table: %v\n", err)
			http.Error(w, "Error querying systems table", http.StatusInternalServerError)
			return
        }
		output += applicationEndpoint + "\n"
    }

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
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
	http.HandleFunc("/view/system", viewSystem)
	http.HandleFunc("/register/topic", registerTopic)
	http.HandleFunc("/subscribe", subscribe)
	http.HandleFunc("/publish", publish)
	log.Printf("Listening for requests...\n")
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		log.Fatalf("Error starting HTTP server: %v\n", err)
	}
}
