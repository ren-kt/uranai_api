package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"text/template"

	"github.com/ren-kt/uranai_api/fortune"
)

const baseUrl = "http://localhost:8080"

type Fortune struct {
	Result string `json:"fortune"`
	Month  string `json:"month"`
	Day    string `json:"day"`
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("views/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, nil)
}

func ResultHandler(w http.ResponseWriter, r *http.Request) {
	month := r.FormValue("month")
	day := r.FormValue("day")

	v := url.Values{"month": {month}, "day": {day}}
	resp, err := http.PostForm(baseUrl+"/api", v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "パラメータが不正です。", http.StatusBadRequest)
		return
	}

	defer resp.Body.Close()

	var f Fortune
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	f.Month = month
	f.Day = day

	t, err := template.ParseFiles("views/result.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, f)
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	month := r.FormValue("month")
	if month == "" {
		http.Error(w, "月が指定されていません", http.StatusBadRequest)
		return
	}

	day := r.FormValue("day")
	if day == "" {
		http.Error(w, "日が指定されていません", http.StatusBadRequest)
		return
	}

	result, err := fortune.GetFortune(fmt.Sprintf("%s%s", month, day))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fortuen := Fortune{Result: result, Month: month, Day: day}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(fortuen); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, buf.String())
}
