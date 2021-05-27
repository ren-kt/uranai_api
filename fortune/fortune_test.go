package fortune_test

import (
	"testing"

	"github.com/ren-kt/uranai_api/fortune"
)

func TestGetFortune(t *testing.T) {
	cases := map[string]struct {
		seed     string
		expected string
		wantErr  bool
	}{
		"11-大吉":          {seed: "11", expected: "大吉", wantErr: false},
		"0101-大吉":        {seed: "0101", expected: "大吉", wantErr: false},
		"1013-中吉":        {seed: "1013", expected: "中吉", wantErr: false},
		"123-吉":          {seed: "123", expected: "吉", wantErr: false},
		"1231-凶":         {seed: "1231", expected: "凶", wantErr: false},
		"empty argument": {seed: "", expected: "", wantErr: true},
		"err2":           {seed: "foo", expected: "", wantErr: true},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			s, err := fortune.GetFortune(tt.seed)
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if s != tt.expected {
				t.Errorf("want %s but got %s", tt.expected, s)
			}
		})
	}
}
