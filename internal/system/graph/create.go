//
// internal/system/graph/create.go
//
package graph

import (
	"gonitorix/internal/graph"
)

func Create() {

	periods := []*graph.GraphPeriod{
		&graph.Daily,
		&graph.Weekly,
		&graph.Monthly,
		&graph.Yearly,
	}

	for _, p := range periods {
		createLoadavg(p)
		createMeminfo(p)
		createProcInfo(p)
		createEntropy(p)
		createUptime(p)
	}
}