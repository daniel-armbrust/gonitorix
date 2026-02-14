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
	"strings"
 )

// SanitizeName converts an arbitrary name into a filesystem-safe string.
// It strips accents, lowercases the input, replaces spaces with dashes,
// and removes any character that is not allowed in filenames.
func SanitizeName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, " ", "_")

	return s
}

// UptimeToString converts a system uptime value in seconds into a
// human-readable string representation.
func UptimeToString(uptime float64) string {
	secs := int64(uptime)

	d := secs / (60 * 60 * 24)
	h := (secs / (60 * 60)) % 24
	m := (secs / 60) % 60

	var dStr string
	if d > 0 {
		dStr = fmt.Sprintf("%d days,", d)
	}

	var result string
	if h > 0 {
		result = fmt.Sprintf("%s %dh %dm", dStr, h, m)
	} else {
		result = fmt.Sprintf("%s %d min", dStr, m)
	}

	return strings.TrimSpace(result)
}