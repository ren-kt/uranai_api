package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "fortune.db")
	if err != nil {
		log.Fatal(err)
	}
	md := NewMyDb(db)

	if err := md.CreateTable(); err != nil {
		log.Fatal(err)
	}

	api := NewApi(http.DefaultClient)

	hs := NewHandlers(md, api)
	_ = hs

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/result", hs.ResultHandler)
	http.HandleFunc("/api", ApiHandler)
	http.HandleFunc("/admin", hs.AdminIndexHandler)
	http.HandleFunc("/admin/create", hs.AdminCreateHandler)
	http.HandleFunc("/admin/edit/", hs.AdminEditHandler)
	http.HandleFunc("/admin/update/", hs.AdminUpdateHandler)
	http.HandleFunc("/admin/delete/", hs.AdminDeleteHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
