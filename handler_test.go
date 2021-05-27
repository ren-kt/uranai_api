package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
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
		month      string
		day        string
		statusCode int
		expected   string
	}{
		"success":             {month: "1", day: "1", statusCode: http.StatusOK, expected: "大吉"},
		"no specifying month": {month: "1", day: "", statusCode: http.StatusBadRequest, expected: "パラメータが不正です。" + "\n"},
		"no specifying day":   {month: "", day: "1", statusCode: http.StatusBadRequest, expected: "パラメータが不正です。" + "\n"},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(ResultHandler))
			defer ts.Close()

			v := url.Values{"month": {tt.month}, "day": {tt.day}}

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
		month      string
		day        string
		statusCode int
		expected   string
	}{
		"success":             {month: "1", day: "1", statusCode: http.StatusOK, expected: `{"fortune":"大吉","month":"1","day":"1"}` + "\n"},
		"no specifying month": {month: "1", day: "", statusCode: http.StatusBadRequest, expected: "日が指定されていません" + "\n"},
		"no specifying day":   {month: "", day: "1", statusCode: http.StatusBadRequest, expected: "月が指定されていません" + "\n"},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(ApiHandler))
			defer ts.Close()

			v := url.Values{"month": {tt.month}, "day": {tt.day}}

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
