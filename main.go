package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/result", ResultHandler)
	http.HandleFunc("/api", ApiHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
