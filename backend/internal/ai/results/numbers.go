package results

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

func boundedNumber(number json.Number) error {
	text := number.String()
	if len(text) > 128 {
		return errBudget
	}
	mantissa := text
	if index := strings.IndexAny(text, "eE"); index >= 0 {
		exponent, err := strconv.ParseInt(text[index+1:], 10, 16)
		if err != nil || exponent < -324 || exponent > 324 {
			return errBudget
		}
		mantissa = text[:index]
	}
	value, err := number.Float64()
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return errJSON
	}
	if value == 0 && strings.ContainsAny(mantissa, "123456789") {
		return errJSON
	}
	return nil
}
