package utils

import (
	"strings"
	"time"
)

func GetDateTimeStringCompatibleWithFileName(date time.Time, formatLayout string) string {
	result := date.Format(formatLayout)
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.ReplaceAll(result, ":", "_")
	result = strings.ReplaceAll(result, "-", "_")
	return result
}
