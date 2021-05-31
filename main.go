package main

import (
	"log"
	"net/http"
)

func main() {
	sqlite, err := NewSqlite()
	if err != nil {
		log.Fatal(err)
	}

	if err := sqlite.CreateTable(); err != nil {
		log.Fatal(err)
	}

	api := NewApi(http.DefaultClient)

	hs := NewHandlers(sqlite, api)

	http.HandleFunc("/", hs.IndexHandler)
	http.HandleFunc("/result", hs.ResultHandler)
	http.HandleFunc("/api", hs.ApiHandler)
	http.HandleFunc("/admin", hs.AdminIndexHandler)
	http.HandleFunc("/admin/create", hs.AdminCreateHandler)
	http.HandleFunc("/admin/edit/", hs.AdminEditHandler)
	http.HandleFunc("/admin/update/", hs.AdminUpdateHandler)
	http.HandleFunc("/admin/delete/", hs.AdminDeleteHandler)
	http.HandleFunc("/admin/upload", hs.AdminUpladHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
