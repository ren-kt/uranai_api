package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/ren-kt/uranai_api/fortune"
)

type Api struct {
	client *http.Client
}

func NewApi(client *http.Client) *Api {
	api := &Api{
		client: client,
	}
	return api
}

const baseURL = "http://localhost:8080"

func (api *Api) Get(month, day int) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api?month=%d&day=%d", baseURL, month, day), nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.request(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("bad response status code %d", resp.StatusCode)
	}

	return resp, err
}

func (api *Api) request(req *http.Request) (*http.Response, error) {
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type Handlers struct {
	md  *MyDb
	api *Api
}

func NewHandlers(md *MyDb, api *Api) *Handlers {
	return &Handlers{md: md, api: api}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("views/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, nil)
}

func (hs Handlers) ResultHandler(w http.ResponseWriter, r *http.Request) {
	month, err := strconv.Atoi(r.FormValue("month"))
	if err != nil || month == 0 {
		http.Error(w, "月が不正です", http.StatusBadRequest)
		return
	}

	day, err := strconv.Atoi(r.FormValue("day"))
	if err != nil || day == 0 {
		http.Error(w, "日が不正です", http.StatusBadRequest)
		return
	}

	resp, err := hs.api.Get(month, day)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var f fortune.Fortune
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

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	month, err := strconv.Atoi(r.FormValue("month"))
	if err != nil || month == 0 {
		fortune := fortune.ApiError{Ok: false, Err: "月が不正なパラメータです"}
		if err := encoder.Encode(fortune); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, buf.String(), http.StatusBadRequest)
		return
	}

	day, err := strconv.Atoi(r.FormValue("day"))
	if err != nil || day == 0 {
		if err != nil || day == 0 {
			fortune := fortune.ApiError{Ok: false, Err: "日が不正なパラメータです"}
			if err := encoder.Encode(fortune); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Error(w, buf.String(), http.StatusBadRequest)
			return
		}
	}

	result, err := fortune.GetFortune(month, day)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO：未実装
	Text := "hoge fuga"

	fortune := fortune.Fortune{Ok: true, Result: result, Text: Text}

	if err := encoder.Encode(fortune); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, buf.String())
}

func (hs *Handlers) AdminIndexHandler(w http.ResponseWriter, r *http.Request) {
	fs, err := hs.md.GetFortuneAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("views/admin/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, fs)
}

func (hs *Handlers) AdminCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		code := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(code), code)
		return
	}

	result := r.FormValue("result")
	if result == "" {
		http.Error(w, "resultが未入力です", http.StatusBadRequest)
		return
	}

	text := r.FormValue("text")
	if text == "" {
		http.Error(w, "textが未入力です", http.StatusBadRequest)
		return
	}

	f := &fortune.Fortune{
		Result: result,
		Text:   text,
	}

	if err := hs.md.Newfortune(f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func (hs *Handlers) AdminEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/admin/edit/"):])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	f, err := hs.md.GetFortune(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("views/admin/edit.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, f)
}

func (hs *Handlers) AdminUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		code := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(code), code)
		return
	}

	id, err := strconv.Atoi(r.URL.Path[len("/admin/update/"):])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := r.FormValue("result")
	if result == "" {
		http.Error(w, "resultが未入力です", http.StatusBadRequest)
		return
	}

	text := r.FormValue("text")
	if text == "" {
		http.Error(w, "textが未入力です", http.StatusBadRequest)
		return
	}

	f := &fortune.Fortune{
		Id:     id,
		Result: result,
		Text:   text,
	}

	if err := hs.md.Updatefortune(f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s%d", "/admin/edit/", id), http.StatusFound)
}

func (hs *Handlers) AdminDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/admin/delete/"):])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := hs.md.Deletefortune(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}
