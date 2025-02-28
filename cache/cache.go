package cache

import (
	"container/heap"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"goudeaux.com/img"
)

var ErrNotFound = errors.New("not found")
var ErrCapacity = errors.New("capacity")

// Entry holds an expirable cached value
type Entry struct {
	Key        string
	Value      string
	Expiration time.Time
	Index      int
}

type internalStats struct {
	Hits      *atomic.Int64
	Misses    *atomic.Int64
	Evictions *atomic.Int64
}

/*
Cache provides a concurrent-safe in-memory key-value cache with expirations.
*/
type Cache struct {
	mutex       *sync.RWMutex
	cache       map[string]*Entry
	expirations *Expirations
	maxSize     int
	stats       internalStats
}

func New(maxSize int) img.Cache {
	return &Cache{
		mutex:       &sync.RWMutex{},
		cache:       map[string]*Entry{},
		expirations: &Expirations{},
		maxSize:     maxSize,
		stats: internalStats{
			Hits:      &atomic.Int64{},
			Misses:    &atomic.Int64{},
			Evictions: &atomic.Int64{}},
	}
}

func (c *Cache) Stats() img.Stats {
	return img.Stats{
		Hits:      c.stats.Hits.Load(),
		Misses:    c.stats.Misses.Load(),
		Evictions: c.stats.Evictions.Load(),
	}
}

// Set adds a key/value to the cache. If at capacity, it deletes
// the expired entry with the oldest expiration. If still at capacity
// it fails.
func (c *Cache) Set(key, value string, exp time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.ensureCapacity()

	entry, present := c.cache[key]
	if present {
		c.updateEntry(entry, value, exp)
	} else {
		if len(c.cache) >= c.maxSize {
			log.Printf("at capacity %s %s", key, value)
			return fmt.Errorf("cannot add key %s: %w", key, ErrCapacity)
		}
		c.addNewEntry(key, value, exp)
	}

	return nil
}

func (c *Cache) addNewEntry(key, value string, exp time.Duration) {
	entry := &Entry{Key: key, Value: value, Expiration: time.Now().Add(exp)}
	c.cache[key] = entry
	heap.Push(c.expirations, entry)

	log.Printf("added %s %s", key, value)
}

func (c *Cache) updateEntry(entry *Entry, value string, exp time.Duration) {
	entry.Value = value
	entry.Expiration = time.Now().Add(exp)
	heap.Fix(c.expirations, entry.Index)
	log.Printf("updated %s %s", entry.Key, value)
}

// Get returns the cached value for a key or returns NotFound
// if the key isn't present or is expired
func (c *Cache) Get(key string) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, ok := c.cache[key]
	if !ok {
		c.stats.Misses.Add(1)
		return "", fmt.Errorf("get key %s: %w", key, ErrNotFound)
	}
	if entry.Expiration.Before(time.Now()) {
		c.stats.Misses.Add(1)
		return "", fmt.Errorf("expired key %s: %w", key, ErrNotFound)
	}
	log.Printf("fetched %s %s", key, entry.Value)
	c.stats.Hits.Add(1)
	return entry.Value, nil
}

// Delete expires a key, staging it for eventual removal
func (c *Cache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, ok := c.cache[key]
	if !ok {
		return fmt.Errorf("delete key %s: %w", key, ErrNotFound)
	}
	entry.Expiration = time.Now().Add(-1 * time.Second)
	heap.Fix(c.expirations, entry.Index)
	log.Printf("deleted %s %s", key, entry.Value)
	return nil
}

func (c *Cache) Prune() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for {
		if len(c.cache) == 0 {
			break
		}

		entry := (*c.expirations)[0]
		if entry.Expiration.After(time.Now()) {
			break
		}
		c.evict(entry)
	}
}

func (c *Cache) ensureCapacity() {
	if len(c.cache) == 0 {
		return
	}

	for len(c.cache) >= c.maxSize {
		entry := (*c.expirations)[0]
		if entry.Expiration.After(time.Now()) {
			break
		}
		c.evict(entry)
	}
}

func (c *Cache) evict(entry *Entry) {
	c.stats.Evictions.Add(1)
	log.Printf("evicted %s %s", entry.Key, entry.Value)
	delete(c.cache, entry.Key)
	_ = heap.Remove(c.expirations, entry.Index)
}
