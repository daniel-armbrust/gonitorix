package utils

import (
	"strconv"
	"math"
	"log"
)

func ParseFloat64(s string) float64 {
	// Parses a string into a float64 value, logging an error on failure, 
	// and rounds the result to 6 decimal places.

    v, err := strconv.ParseFloat(s, 64)

    if err != nil {
        log.Printf("failed to parse float value %q", s)
        return 0
    }

    return math.Round(v*1e6) / 1e6
}

func RoundFloat64(v float64) float64 {
	// Rounds a float64 value to a fixed precision of 6 decimal places.
    return math.Round(v*1e6) / 1e6
}