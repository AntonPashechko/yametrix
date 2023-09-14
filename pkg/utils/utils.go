// Package utils содержит методы конвертирования типов данных.
package utils

import (
	"fmt"
	"strconv"
)

// StrToInt64 Просто переводит строку в int64 с проверкой ошибки
func StrToInt64(str string) (int64, error) {
	if str != "" {
		return strconv.ParseInt(str, 10, 64)
	}
	return 0, nil
}

// StrToFloat64 Просто переводит строку в float64 с проверкой ошибки
func StrToFloat64(str string) (float64, error) {
	if str != "" {
		return strconv.ParseFloat(str, 64)
	}
	return 0, nil
}

// Int64ToStr Просто переводит float64 в строку
func Int64ToStr(i64 int64) string {
	return fmt.Sprintf("%d", i64)
}

// Float64ToStr Просто переводит float64 в строку
func Float64ToStr(f64 float64) string {
	return strconv.FormatFloat(f64, 'f', -1, 64)
}
