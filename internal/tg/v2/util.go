package tg

import (
	"strconv"
	"time"
)

func itoa(i int) string {
	return strconv.Itoa(i)
}

func seconds(i int) time.Duration {
	return time.Duration(i) * time.Second
}

// formMap
type fM = map[string]string

// jsonMap
type jM = map[string]any
