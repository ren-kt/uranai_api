package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Fortune struct {
	Result string `json:"fortune"`
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
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

	result, err := getFortune(fmt.Sprintf("%s%s", month, day))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fortuen := Fortune{Result: result}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(fortuen); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, buf.String())
}

func getFortune(date string) (string, error) {
	var seed int
	for _, s := range strings.Split(date, "") {
		i, err := strconv.Atoi(s)
		if err != nil {
			return "", err
		}
		seed += i
	}

	if seed >= 10 {
		var tmp_seed int
		for _, s := range strings.Split(strconv.Itoa(seed), "") {
			i, err := strconv.Atoi(s)
			if err != nil {
				return "", err
			}
			tmp_seed += i
		}
		seed = tmp_seed
	}

	var fortune string
	switch seed {
	case 2:
		fortune = "大吉"
	case 1, 5:
		fortune = "中吉"
	case 3, 6, 8:
		fortune = "吉"
	case 4, 7, 9:
		fortune = "凶"
	default:
		fortune = "吉"
	}

	return fortune, nil
}
