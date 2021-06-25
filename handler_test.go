package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/ren-kt/uranai_api/fortune"
)

type TestDB struct{}

func (d *TestDB) CreateTable() error {
	return nil
}

func (d *TestDB) GetText(result string) (string, error) {
	return "test text", nil
}

func (d *TestDB) GetFortune(id int) (*fortune.Fortune, error) {
	return nil, nil
}

func (d *TestDB) GetFortuneAll() ([]*fortune.Fortune, error) {
	return nil, nil
}

func (d *TestDB) Updatefortune(f *fortune.Fortune) error {
	return nil
}

func (d *TestDB) Deletefortune(id int) error {
	return nil
}

func (d *TestDB) Newfortune(fortune *fortune.Fortune) error {
	return nil
}

func (d *TestDB) MultipleNewfortune(entityCh <-chan []string, multipluNum int) <-chan error {
	return nil
}

var _ DB = &TestDB{}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func client(t *testing.T, day, month int) *http.Client {
	t.Helper()

	result, _ := fortune.GetFortune(month, day)
	body := fortune.Fortune{Result: result}

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	return NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(b)),
			Header:     make(http.Header),
		}
	})
}

func TestIndexHandler(t *testing.T) {
	cases := map[string]struct {
		statusCode int
	}{
		"success": {statusCode: http.StatusOK},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			hs := NewHandlers(nil, nil)
			ts := httptest.NewServer(http.HandlerFunc(hs.IndexHandler))
			defer ts.Close()

			resp, err := http.Get(ts.URL)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("unexpected status code: %d", resp.StatusCode)
			}
		})
	}
}

func TestResultHandler(t *testing.T) {
	cases := map[string]struct {
		month      int
		day        int
		statusCode int
		expected   string
	}{
		"success":             {month: 1, day: 1, statusCode: http.StatusOK, expected: "大吉"},
		"no specifying month": {month: 1, day: 0, statusCode: http.StatusBadRequest, expected: "日が不正です" + "\n"},
		"no specifying day":   {month: 0, day: 1, statusCode: http.StatusBadRequest, expected: "月が不正です" + "\n"},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			api := NewApi(client(t, tt.month, tt.day))
			hs := NewHandlers(nil, api)

			ts := httptest.NewServer(http.HandlerFunc(hs.ResultHandler))
			defer ts.Close()

			v := url.Values{"month": {strconv.Itoa(tt.month)}, "day": {strconv.Itoa(tt.day)}}

			resp, err := http.PostForm(ts.URL, v)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("unexpected status code: %d", resp.StatusCode)
			}

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}

			if !strings.Contains(string(b), tt.expected) {
				t.Errorf("unexpected response: %s cannot find %s", tt.expected, string(b))
			}
		})
	}
}

func TestApiHandler(t *testing.T) {
	cases := map[string]struct {
		month      int
		day        int
		statusCode int
		expected   string
	}{
		"success":             {month: 1, day: 1, statusCode: http.StatusOK, expected: `{"ok":true,"resut":"大吉","text":"test text"}` + "\n"},
		"no specifying month": {month: 1, day: 0, statusCode: http.StatusBadRequest, expected: `{"ok":false,"error":"日が不正なパラメータです"}` + "\n\n"},
		"no specifying day":   {month: 0, day: 1, statusCode: http.StatusBadRequest, expected: `{"ok":false,"error":"月が不正なパラメータです"}` + "\n\n"},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			td := &TestDB{}
			hs := NewHandlers(td, nil)

			ts := httptest.NewServer(http.HandlerFunc(hs.ApiHandler))
			defer ts.Close()

			v := url.Values{"month": {strconv.Itoa(tt.month)}, "day": {strconv.Itoa(tt.day)}}

			resp, err := http.PostForm(ts.URL, v)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("unexpected status code: %d", resp.StatusCode)
			}

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}

			if s := string(b); s != tt.expected {
				t.Errorf("unexpected response: %s", s)
			}
		})
	}
}

func TestAdminIndexHandler(t *testing.T) {
	cases := map[string]struct {
		statusCode int
	}{
		"success": {statusCode: http.StatusOK},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			td := &TestDB{}
			hs := NewHandlers(td, nil)
			ts := httptest.NewServer(http.HandlerFunc(hs.AdminIndexHandler))
			defer ts.Close()

			resp, err := http.Get(ts.URL)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("unexpected status code: %d", resp.StatusCode)
			}
		})
	}
}

func TestAdminCreateHandler(t *testing.T) {
	cases := map[string]struct {
		result     string
		text       string
		statusCode int
	}{
		"success":                             {result: "大吉", text: "test text", statusCode: http.StatusOK},
		"error with missing result parameter": {result: "", text: "test text", statusCode: http.StatusBadRequest},
		"error with missing text parameter":   {result: "大吉", text: "", statusCode: http.StatusBadRequest},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			td := &TestDB{}
			hs := NewHandlers(td, nil)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/admin" {
					hs.AdminIndexHandler(w, r)
				} else {
					hs.AdminCreateHandler(w, r)
				}
			}))
			defer ts.Close()

			v := url.Values{"result": {tt.result}, "text": {tt.text}}
			resp, err := http.PostForm(ts.URL, v)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("unexpected status code: %d", resp.StatusCode)
			}
		})
	}
}

func TestAdminEditHandler(t *testing.T) {
	cases := map[string]struct {
		id         string
		statusCode int
	}{
		"success":                       {id: "1", statusCode: http.StatusOK},
		"error where id is a character": {id: "a", statusCode: http.StatusInternalServerError},
		"error where id is empty":       {id: "", statusCode: http.StatusInternalServerError},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			td := &TestDB{}
			hs := NewHandlers(td, nil)
			ts := httptest.NewServer(http.HandlerFunc(hs.AdminEditHandler))
			defer ts.Close()

			resp, err := http.Get(fmt.Sprintf("%s%s%s", ts.URL, "/admin/edit/", tt.id))
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("unexpected status code: %d", resp.StatusCode)
			}
		})
	}
}

func TestAdminUpdateHandler(t *testing.T) {
	cases := map[string]struct {
		id         string
		result     string
		text       string
		statusCode int
	}{
		"success":                             {id: "1", result: "大吉", text: "test text", statusCode: http.StatusOK},
		"error with missing result parameter": {id: "1", result: "", text: "test text", statusCode: http.StatusBadRequest},
		"error with missing text parameter":   {id: "1", result: "大吉", text: "", statusCode: http.StatusBadRequest},
		"error where id is a character":       {id: "a", result: "大吉", text: "test text", statusCode: http.StatusInternalServerError},
		"error where id is empty":             {id: "", result: "大吉", text: "test text", statusCode: http.StatusInternalServerError},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			td := &TestDB{}
			hs := NewHandlers(td, nil)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == fmt.Sprintf("%s%s", "/admin/edit/", tt.id) {
					hs.AdminEditHandler(w, r)
				} else {
					hs.AdminUpdateHandler(w, r)
				}
			}))
			defer ts.Close()

			v := url.Values{"result": {tt.result}, "text": {tt.text}}
			resp, err := http.PostForm(fmt.Sprintf("%s%s%s", ts.URL, "/admin/update/", tt.id), v)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("unexpected status code: %d", resp.StatusCode)
			}
		})
	}
}

func TestAdminDeleteHandler(t *testing.T) {
	cases := map[string]struct {
		id         string
		statusCode int
	}{
		"success":                             {id: "1", statusCode: http.StatusOK},
		"error with missing result parameter": {id: "a", statusCode: http.StatusInternalServerError},
		"error with missing text parameter":   {id: "", statusCode: http.StatusInternalServerError},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			td := &TestDB{}
			hs := NewHandlers(td, nil)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/admin" {
					hs.AdminIndexHandler(w, r)
				} else {
					hs.AdminDeleteHandler(w, r)
				}
			}))
			defer ts.Close()

			resp, err := http.Get(fmt.Sprintf("%s%s%s", ts.URL, "/admin/delete/", tt.id))
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("unexpected status code: %d", resp.StatusCode)
			}
		})
	}
}
