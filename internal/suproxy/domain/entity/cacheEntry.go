package entity

import (
	"encoding/json"
	"time"
)

// CacheEntry represents how a single cache item is stored in Redis
type CacheEntry struct {
	Mock     bool            `json:"mock"`
	Key      string          `json:"key"`
	Request  Request         `json:"request"`
	Response json.RawMessage `json:"response"`
	CachedAt time.Time       `json:"cached_at"`
	Version  int             `json:"v"`
}
