package handler

import (
	"errors"
	"fmt"
	"github.com/martinlindhe/base36"
	"strconv"
	"strings"
)

func promoStep(input string, forward bool) (string, error) {
	split := strings.Split(input, "-")
	// 2 - parts of separated
	// 4 - promo length after minus cymbol

	promo := split[1]
	if len(split) != 2 || len(promo) != 4 {
		return "", errors.New("invalid promo")
	}

	if !forward {
		check, err := strconv.Atoi(promo)
		if err == nil && check == 0 {
			return "", errors.New("code has the lowest border")
		}
	}

	if len(promo) != lengthCode {
		return "", errors.New(fmt.Sprintf("length of code after '-' cymbol must be %d for the event", lengthCode))
	}

	decode := base36.Decode("1" + promo)

	if forward {
		decode++
	} else {
		decode--
	}

	encode := base36.Encode(decode)

	return preset + strings.ToLower(encode[1:]), nil
}
