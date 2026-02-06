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
	"fmt"
	"math"
)

const (
	DaySeconds   = 86400
	WeekSeconds  = 7 * DaySeconds
	MonthSeconds = 31 * DaySeconds
	YearSeconds  = 365 * DaySeconds
)

// Heartbeat returns a safe heartbeat value for a given RRD step.
// The rule used is: heartbeat = step * 2.
func Heartbeat(step int) int {
	return step * 2
}

// Rows calculates how many rows an RRA needs to store "durationSeconds"
// when each row represents pdpPerRow primary data points.
func Rows(step, pdpPerRow, durationSeconds int) int {
	return durationSeconds / (step * pdpPerRow)
}

// RRA builds a formatted RRA string.
func RRA(cf string, xff float64, pdpPerRow, rows int) string {
	return fmt.Sprintf(
		"RRA:%s:%.1f:%d:%d",
		cf,
		xff,
		pdpPerRow,
		rows,
	)
}

// Converts float values into a format accepted by rrdtool,
// returning "U" for NaN or infinite and formatting valid numbers otherwise.
func RRDfloat(v float64, prec int) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "U"
	}

	format := fmt.Sprintf("%%.%df", prec)

	return fmt.Sprintf(format, v)
}