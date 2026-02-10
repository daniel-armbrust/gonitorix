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

package logging

import (
	"log"
	"sync/atomic"
)

type Level int32

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var currentLevel atomic.Int32

func SetDebug(enabled bool) {
	if enabled {
		currentLevel.Store(int32(LevelDebug))
	} else {
		currentLevel.Store(int32(LevelInfo))
	}
}

func DebugEnabled() bool {
	return Level(currentLevel.Load()) == LevelDebug
}

func logf(prefix string, level Level, tag string, format string, args ...any) {
	if level < Level(currentLevel.Load()) {
		return
	}

	log.Printf("[%s][%s] "+format, append([]any{prefix, tag}, args...)...)
}

func Debug(pkg string, format string, args ...any) {
	logf(pkg, LevelDebug, "DEBUG", format, args...)
}

func Info(pkg string, format string, args ...any) {
	logf(pkg, LevelInfo, "INFO", format, args...)
}

func Warn(pkg string, format string, args ...any) {
	logf(pkg, LevelWarn, "WARN", format, args...)
}

func Error(pkg string, format string, args ...any) {
	logf(pkg, LevelError, "ERROR", format, args...)
}
