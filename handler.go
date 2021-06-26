package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/ren-kt/uranai_api/fortune"
	"golang.org/x/sync/errgroup"
)

const baseURL = "http://localhost:8080"

type Api struct {
	client *http.Client
}

func NewApi(client *http.Client) *Api {
	api := &Api{
		client: client,
	}
	return api
}

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
	db                  DB
	api                 *Api
	singleProcessTime   time.Duration
	multipleProcessTime time.Duration
}

func NewHandlers(db DB, api *Api) *Handlers {
	return &Handlers{db: db, api: api}
}

func (hs Handlers) IndexHandler(w http.ResponseWriter, r *http.Request) {
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

func (hs Handlers) ApiHandler(w http.ResponseWriter, r *http.Request) {
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

	text, err := hs.db.GetText(result)
	if err == sql.ErrNoRows {
		fortune := fortune.ApiError{Ok: false, Err: "textが見つかりません"}
		if err := encoder.Encode(fortune); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, buf.String(), http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fortune := fortune.Fortune{Ok: true, Result: result, Text: text}

	if err := encoder.Encode(fortune); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, buf.String())
}

func (hs *Handlers) AdminIndexHandler(w http.ResponseWriter, r *http.Request) {
	fs, err := hs.db.GetFortuneAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("views/admin/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Fortunes            []*fortune.Fortune
		SingleProcessTime   time.Duration
		MultipleProcessTime time.Duration
	}{
		Fortunes:            fs,
		SingleProcessTime:   hs.singleProcessTime,
		MultipleProcessTime: hs.multipleProcessTime,
	}

	t.Execute(w, data)
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

	if err := hs.db.Newfortune(f); err != nil {
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

	f, err := hs.db.GetFortune(id)
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

	if err := hs.db.Updatefortune(f); err != nil {
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

	if err := hs.db.Deletefortune(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

// 19.4148119s 並列数1  10000row
func (hs *Handlers) AdminUpladHandler(w http.ResponseWriter, r *http.Request) {
	t1 := time.Now()

	if r.Method != http.MethodPost {
		code := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(code), code)
		return
	}

	file, _, err := r.FormFile("uploaded")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	_, err = reader.Read()
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fortune := &fortune.Fortune{Result: line[0], Text: line[1]}
		err = hs.db.Newfortune(fortune)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	t2 := time.Now()
	hs.singleProcessTime = t2.Sub(t1)

	http.Redirect(w, r, "/admin", http.StatusFound)
}

// 13.2647466s 並列数1  10000row
// 9.8293747s  並列数2  10000row
// 7.8937453s  並列数3  10000row
// 6.2955198s  並列数4  10000row
// 4.8378591s  並列数10 10000row
func (hs *Handlers) AdminMultipleUpladHandler(w http.ResponseWriter, r *http.Request) {
	t1 := time.Now()

	if r.Method != http.MethodPost {
		code := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(code), code)
		return
	}

	file, _, err := r.FormFile("uploaded")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	defer file.Close()

	multipluNum, err := strconv.Atoi(r.FormValue("multiple"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reader := csv.NewReader(file)
	_, err = reader.Read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lineCh := make(chan []string)
	var eg errgroup.Group
	m := new(sync.Mutex)
	eg.Go(func() error {
		var returnErr error
		for {
			line, err := reader.Read()
			if err == io.EOF {
				close(lineCh)
				break
			} else if err != nil {
				close(lineCh)
				returnErr = err
				break
			}
			m.Lock()
			lineCh <- line
			m.Unlock()
		}
		return returnErr
	})

	if err = <-hs.db.MultipleNewfortune(lineCh, multipluNum); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := eg.Wait(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t2 := time.Now()
	hs.multipleProcessTime = t2.Sub(t1)

	http.Redirect(w, r, "/admin", http.StatusFound)
}
