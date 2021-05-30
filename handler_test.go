package main

import (
	"bytes"
	"encoding/json"
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

func TestIndexHandler(t *testing.T) {
	cases := map[string]struct {
		statusCode int
	}{
		"success": {statusCode: http.StatusOK},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(IndexHandler))
			defer ts.Close()

			resp, err := http.Get(ts.URL)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("unexpected status code: %d", resp.StatusCode)
			}

			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("unexpected error %s", err)
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
		"success":             {month: 1, day: 1, statusCode: http.StatusOK, expected: `{"ok":true,"resut":"大吉","text":"hoge fuga"}` + "\n"},
		"no specifying month": {month: 1, day: 0, statusCode: http.StatusBadRequest, expected: `{"ok":false,"error":"日が不正なパラメータです"}` + "\n\n"},
		"no specifying day":   {month: 0, day: 1, statusCode: http.StatusBadRequest, expected: `{"ok":false,"error":"月が不正なパラメータです"}` + "\n\n"},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(ApiHandler))
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
