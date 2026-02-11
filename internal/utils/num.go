/*
 * Gonitorix - a system and network monitoring tool
 * Copyright (C) 2026 Daniel Armbrust <darmbrust@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
 
package utils

import (
	"strconv"
	"math"
	
	"gonitorix/internal/logging"
)

// ParseFloat64 converts a numeric string into float64.
// It returns 0 when the value cannot be parsed.
func ParseFloat64(s string) float64 {
	// Parses a string into a float64 value, logging an error on failure, 
	// and rounds the result to 6 decimal places.

    v, err := strconv.ParseFloat(s, 64)

    if err != nil {
        logging.Error("NETIF", "Failed to parse float value %q\n", s)
        return 0
    }

    return math.Round(v*1e6) / 1e6
}

// RoundFloat64 rounds a float64 value to the specified number of decimal 
// places.
func RoundFloat64(v float64) float64 {
	// Rounds a float64 value to a fixed precision of 6 decimal places.
    return math.Round(v*1e6) / 1e6
}