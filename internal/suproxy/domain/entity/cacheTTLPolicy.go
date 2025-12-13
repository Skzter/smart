package entity

import "time"

// CacheTTLPolicy defines TTL durations for different cache cases.
type CacheTTLPolicy struct {
	SupplierOK   time.Duration
	MockOK       time.Duration
	ErrorOrEmpty time.Duration
}
