package fortune

import (
	"fmt"
	"strconv"
	"strings"
)

type Fortune struct {
	Id     int    `json:"-"`
	Ok     bool   `json:"ok"`
	Result string `json:"resut"`
	Text   string `json:"text"`
	Month  int    `json:"-"`
	Day    int    `json:"-"`
}

type ApiError struct {
	Ok  bool   `json:"ok"`
	Err string `json:"error"`
}

func GetFortune(month, day int) (string, error) {
	date := fmt.Sprintf("%d%d", month, day)
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
	}

	return fortune, nil
}
