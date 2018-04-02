package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

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
	Data interface{} `json:"data"`
}

type EventLog struct {
	Topic string
	Success bool
	CreatedAt time.Time
}

type TemplateArgs struct {
	Systems []System
	Topics []Topic
	SuccessEvents map[string]map[time.Time]int
	FailureEvents map[string]map[time.Time]int
}

func minusMonth(month time.Month) int {
	return int(month)
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
    	http.Error(w, "404 not found", http.StatusNotFound)
        return
    } else if r.Method != "GET" {
		http.Error(w, "Only GET methods are supported", http.StatusBadRequest)
		return
	}

	systemRows, err := db.Query(fmt.Sprintf("SELECT * FROM systems"))
	if err != nil {
		log.Printf("Error querying systems table: %v\n", err)
    	http.Error(w, "Error querying systems table", http.StatusInternalServerError)
		return
    }
	defer systemRows.Close()

	var allSystemEntries []System
	for systemRows.Next() {
		var system System

        if err := systemRows.Scan(&system.Name, &system.ApplicationEndpoint); err != nil {
			log.Printf("Error querying systems table: %v\n", err)
			http.Error(w, "Error querying systems table", http.StatusInternalServerError)
			return
		}
		
		allSystemEntries = append(allSystemEntries, system)
	}

	topicRows, err := db.Query(fmt.Sprintf("SELECT * FROM topics"))
	if err != nil {
		log.Printf("Error querying topics table: %v\n", err)
    	http.Error(w, "Error querying topics table", http.StatusInternalServerError)
		return
    }
	defer topicRows.Close()

	var allTopicEntries []Topic
	for topicRows.Next() {
		var topic Topic

        if err := topicRows.Scan(&topic.Name, &topic.Description, &topic.Owner, &topic.Structure, pq.Array(&(topic.Subscribers))); err != nil {
			log.Printf("Error querying topics table: %v\n", err)
			http.Error(w, "Error querying topics table", http.StatusInternalServerError)
			return
		}
		
		allTopicEntries = append(allTopicEntries, topic)
	}

	eventLogRows, err := db.Query(fmt.Sprintf("SELECT * FROM events"))
	if err != nil {
		log.Printf("Error querying events table: %v\n", err)
    	http.Error(w, "Error querying events table", http.StatusInternalServerError)
		return
    }
	defer eventLogRows.Close()

	var allEventLogs []EventLog
	for eventLogRows.Next() {
		var eventLog EventLog

        if err := eventLogRows.Scan(&eventLog.Topic, &eventLog.Success, &eventLog.CreatedAt); err != nil {
			log.Printf("Error querying events table: %v\n", err)
			http.Error(w, "Error querying events table", http.StatusInternalServerError)
			return
		}
		
		allEventLogs = append(allEventLogs, eventLog)
	}

	//indexTemplate, err := template.New("").Funcs(template.FuncMap{"minusMonth": minusMonth}).ParseFiles("index.html")
	indexTemplate, err := template.ParseFiles("index.html")
	if err != nil {
		log.Printf("Error parsing index.html: %v\n", err)
		http.Error(w, "Error parsing index.html", http.StatusInternalServerError)
		return
	}
	templateArgs := TemplateArgs{Systems: allSystemEntries, Topics: allTopicEntries}
	templateArgs.SuccessEvents = make(map[string]map[time.Time]int)
	templateArgs.FailureEvents = make(map[string]map[time.Time]int)
	for _, eventLog := range allEventLogs {
		t := time.Date(eventLog.CreatedAt.Year(), eventLog.CreatedAt.Month(), eventLog.CreatedAt.Day(), eventLog.CreatedAt.Hour(), eventLog.CreatedAt.Minute(), eventLog.CreatedAt.Second(), 0, eventLog.CreatedAt.Location())
		if eventLog.Success {
			if templateArgs.SuccessEvents[eventLog.Topic] == nil {
				templateArgs.SuccessEvents[eventLog.Topic] = make(map[time.Time]int)
			}
			
			templateArgs.SuccessEvents[eventLog.Topic][t] += 1
		} else {
			if templateArgs.FailureEvents[eventLog.Topic] == nil {
				templateArgs.FailureEvents[eventLog.Topic] = make(map[time.Time]int)
			}

			templateArgs.FailureEvents[eventLog.Topic][t] += 1
		}
	}
	indexTemplate.Execute(w, templateArgs)
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
		log.Printf("Error decoding published topic event POST: %v\n", err)
    	http.Error(w, "Error decoding POST body", http.StatusBadRequest)
		return
	}
	log.Printf("%+v\n", newEvent)

	_, err = json.Marshal(&(newEvent.Data))
	if (err != nil) {
		log.Printf("Invalid JSON data in published event: %v\n", err)
    	http.Error(w, "Invalid JSON in data field", http.StatusBadRequest)
		return
	}

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
	
	subscriberApplicationEndpointRows, err := db.Query(fmt.Sprintf("SELECT applicationEndpoint FROM systems WHERE name in (%v)", subscriptionSearchString))
	if err != nil {
		log.Printf("Error querying systems table: %v\n", err)
    	http.Error(w, "Error querying systems table", http.StatusInternalServerError)
		return
    }
	defer subscriberApplicationEndpointRows.Close()

	output := "Attempted publishing to following endpoints:\n"
	for subscriberApplicationEndpointRows.Next() {
		var applicationEndpoint string
        if err := subscriberApplicationEndpointRows.Scan(&applicationEndpoint); err != nil {
			log.Printf("Error scanning system table: %v\n", err)
			continue
        }
		go sendPost(applicationEndpoint, &newEvent)
		output += applicationEndpoint + "\n"
    }

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
}

func sendPost(url string, data *Event) {
	jsonData := new(bytes.Buffer)
	err := json.NewEncoder(jsonData).Encode(data)
	if err != nil {
		log.Printf("Error encoding data for URL '%v': %v\nData being encoded was:\n%+v", url, err, *data)
		logEvent(data.Topic, false)
		return
	}

	req, err := http.NewRequest("POST", url, jsonData)
	if err != nil {
		log.Printf("Error creating request for URL '%v': %v\nData being sent was:\n%+v", url, err, *data)
		logEvent(data.Topic, false)
		return
	}
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: time.Minute}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error posting data for URL '%v': %v\nData being sent was:\n%+v", url, err, *data)
		logEvent(data.Topic, false)
		return
    }
    defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoder := json.NewDecoder(resp.Body)
		var respBody string
		err := decoder.Decode(&respBody)
		if err != nil {
        	log.Printf("Error decoding failure response body for URL '%v' with status '%v': %v\nData being sent was:\n%+v", url, resp.Status, err, *data)
		} else {
			log.Printf("Error posting data for URL '%v' with status '%v': %v\nData being sent was:\n%+v", url, resp.Status, respBody, *data)
		}
		logEvent(data.Topic, false)
		return
	}

	log.Printf("Successfully sent POST to URL '%v'\nData being sent was:\n%+v", url, *data)
	logEvent(data.Topic, true)
}

func logEvent(topicName string, success bool) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS events (name TEXT NOT NULL, success BOOLEAN NOT NULL, created_at TIMESTAMPTZ DEFAULT now() NOT NULL)")
	if err != nil {
		log.Printf("Error creating events table: %v\n", err)
		return
    }

	_, err = db.Exec(fmt.Sprintf("INSERT INTO events (name, success) VALUES ('%v', %v)", topicName, success))
	if err != nil {
		log.Printf("Error creating new events row: %v\n", err)
		return
    }
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
