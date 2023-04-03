package utils

import (
	"strconv"
	"time"
)

func GetCurrentDate() string {
	today := time.Now().UTC()

	var (
		month string
		day   string
	)

	monthNum := int(today.Month())
	if monthNum < 10 {
		month = "0" + strconv.Itoa(monthNum)
	} else {
		month = strconv.Itoa(monthNum)
	}

	dayNum := today.Day()
	if dayNum < 10 {
		day = "0" + strconv.Itoa(dayNum)
	} else {
		day = strconv.Itoa(dayNum)
	}

	year := strconv.Itoa(today.Year())

	return year + "-" + month + "-" + day
}
