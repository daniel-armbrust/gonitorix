//
// internal/net/globals.go
//
package net

var (
	// lastTimestamp stores the timestamp of the previous polling cycle and is
	// used to compute the elapsed time (deltaT) between two reads.
	lastTimestamp float64

	// lastIfStats stores the previous statistics snapshot for each network
	// interface, used to compute deltas between successive polling cycles.
	lastIfstats = make(map[string]ifStats)
)