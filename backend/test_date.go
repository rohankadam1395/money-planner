//go:build ignore

package main

import (
	"fmt"
	"time"
)

func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"02-01-2006",
		"01/02/2006",
		"2006/01/02",
		"January 2, 2006",
		"02 Jan 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func main() {
	dates := []string{"22/07/2026", "01/07/2026", "2026-01-07"}
	for _, d := range dates {
		t, err := parseDate(d)
		fmt.Printf("%s: %v, err=%v\n", d, t, err)
	}
}
