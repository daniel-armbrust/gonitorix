//
// internal/net/structs.go
//
package net

// ifStats holds the raw network interface counters read from /proc/net/dev
// for a single polling cycle.
type ifStats struct {	 
	rxBytes  float64
	txBytes  float64	
	rxPkts   float64
	txPkts   float64
	rxErrors float64
	txErrors float64
}