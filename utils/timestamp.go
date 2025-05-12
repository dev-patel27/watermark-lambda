package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ExtractTimestamp(filename string) (time.Time, error) {
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid filename format: %s", filename)
	}
	tsStr := strings.TrimSuffix(parts[1], ".mp4")
	ms, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp in filename: %s", tsStr)
	}
	return time.Unix(0, ms*int64(time.Millisecond)), nil
}
