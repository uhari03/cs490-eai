package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"strings"
)

func sayHello(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message
	w.Write([]byte(message))
}

func main() {
	port := os.Getenv("PORT")

	http.HandleFunc("/", sayHello)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Fatal("$PORT must be set")
	}
}
