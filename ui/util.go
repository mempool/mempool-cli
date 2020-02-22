package ui

import (
	"fmt"
	"math"
)

func ceil(f float64) int {
	return int(
		math.Ceil(f),
	)
}

func fmtSeconds(s int64) string {
	if s < 60 {
		return "< 1 minute"
	} else if s < 120 {
		return fmt.Sprintf("1 minute")
	} else if s < 3600 {
		return fmt.Sprintf("%d minutes", s/60)
	}
	return fmt.Sprintf("%d hours", s/3600)
}

func fmtSize(s int) string {
	if s < 1024*1024 {
		m := float64(s) / 1000.0
		return fmt.Sprintf("%dkB", ceil(m))
	}
	m := float64(s) / (1000.0 * 1000.0)
	return fmt.Sprintf("%dMB", ceil(m))
}
