package utils

import (
	"strconv"
)

/*Просто переводит строку в int64 с проверкой ошибки*/
func StrToInt64(str string) (int64, error) {
	if str != "" {
		return strconv.ParseInt(str, 10, 64)
	}
	return 0, nil
}

/*Просто переводит строку в float64 с проверкой ошибки*/
func StrToFloat64(str string) (float64, error) {
	if str != "" {
		return strconv.ParseFloat(str, 64)
	}
	return 0, nil
}
