package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/result", resultHandler)
	http.HandleFunc("/api", apiHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
