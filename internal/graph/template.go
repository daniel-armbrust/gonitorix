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
 
package graph

import (
	"strconv"

    "gonitorix/internal/config"
)

func BuildGraphArgs(t GraphTemplate) []string {
    // Builds the rrdtool argument list.

    args := []string{
        "graph", t.Graph,
        "--title=" + t.Title,
        "--start=" + t.Start,
        "--imgformat=PNG",
        "--vertical-label=" + t.VerticalLabel,
        "--width=" + strconv.Itoa(config.GlobalCfg.GraphWidth),
        "--height=" + strconv.Itoa(config.GlobalCfg.GraphHeight),
        "--full-size-mode",
        "--zoom=1",
        "--slope-mode",
        "--font=LEGEND:7:",
        "--font=TITLE:9:",
        "--font=UNIT:8:",
        "--font=DEFAULT:0:Mono",
    }

    if t.XGrid != "" {
        args = append(args, "--x-grid=" + t.XGrid)
    }

    args = append(args,
        "--color=CANVAS#000000",
        "--color=BACK#101010",
        "--color=FONT#C0C0C0",
        "--color=MGRID#80C080",
        "--color=GRID#808020",
        "--color=FRAME#808080",
        "--color=ARROW#FFFFFF",
        "--color=SHADEA#404040",
        "--color=SHADEB#404040",
        "--color=AXIS#101010",
    )

    args = append(args, t.Defs...)
    args = append(args, t.CDefs...)
    args = append(args, t.Draw...)

    return args
}