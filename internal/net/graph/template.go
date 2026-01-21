package graph

import (
	"strconv"
)

type graphTemplate struct {
    img           string
    title         string
    start         string
    verticalLabel string
    width         int
    height        int
    xGrid         string
    defs          []string
    cdefs         []string
    draw          []string
}

func buildGraphArgs(t graphTemplate) []string {
    // Builds the rrdtool argument list required to generate a 
    // network interface graph.

    args := []string{
        "graph", t.img,
        "--title=" + t.title,
        "--start=" + t.start,
        "--imgformat=PNG",
        "--vertical-label=" + t.verticalLabel,
        "--width=" + itoa(t.width),
        "--height=" + itoa(t.height),
        "--zoom=1",
        "--slope-mode",
        "--font=LEGEND:7:",
        "--font=TITLE:9:",
        "--font=UNIT:8:",
        "--font=DEFAULT:0:Mono",
    }

    if t.xGrid != "" {
        args = append(args, "--x-grid=" + t.xGrid)
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

    args = append(args, t.defs...)
    args = append(args, t.cdefs...)
    args = append(args, t.draw...)

    return args
}

func itoa(v int) string {
    return strconv.Itoa(v)
}
