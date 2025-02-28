package img

import "time"

// Stats provides a snapshot of cache history since instantiation
type Stats struct {
	Hits      int64
	Misses    int64
	Evictions int64
}

type Cache interface {
	Set(string, string, time.Duration) error
	Get(string) (string, error)
	Delete(string) error
	Stats() Stats
	Prune()
}
