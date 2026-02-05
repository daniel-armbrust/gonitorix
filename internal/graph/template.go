//
// internal/graph/template.go
//
package graph

import (
	"strconv"
)

func BuildGraphArgs(t GraphTemplate) []string {
    // Builds the rrdtool argument list.

    args := []string{
        "graph", t.Graph,
        "--title=" + t.Title,
        "--start=" + t.Start,
        "--imgformat=PNG",
        "--vertical-label=" + t.VerticalLabel,
        "--width=" + strconv.Itoa(t.Width),
        "--height=" + strconv.Itoa(t.Height),
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