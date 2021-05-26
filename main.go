package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api", apiHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
