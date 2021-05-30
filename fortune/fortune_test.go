package fortune_test

import (
	"testing"

	"github.com/ren-kt/uranai_api/fortune"
)

func TestGetFortune(t *testing.T) {
	cases := map[string]struct {
		month    int
		day      int
		expected string
		wantErr  bool
	}{
		"11-大吉":          {month: 1, day: 1, expected: "大吉", wantErr: false},
		"1013-中吉":        {month: 10, day: 13, expected: "中吉", wantErr: false},
		"123-吉":          {month: 1, day: 23, expected: "吉", wantErr: false},
		"1231-凶":         {month: 12, day: 31, expected: "凶", wantErr: false},
		"911-凶":         {month: 9, day: 11, expected: "大吉", wantErr: false},
		"empty argument": {month: 0, day: 0, expected: "", wantErr: true},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			s, err := fortune.GetFortune(tt.month, tt.day)
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if s != tt.expected {
				t.Errorf("want %s but got %s", tt.expected, s)
			}
		})
	}
}
