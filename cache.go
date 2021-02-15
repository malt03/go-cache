package cache

import (
	"math/rand"
	"sync"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// Cache caches a single value. To cache multiple values, the same number of instances is required. It is thread-safe.
type Cache struct {
	value     interface{}
	expiresAt time.Time
	config    *Config
	lock      sync.Mutex
}

// New constructs a Cache.
func New(config *Config) *Cache {
	return &Cache{
		value:     nil,
		expiresAt: time.Now(),
		config:    config,
	}
}

// Config is a structure that is passed to New.
type Config struct {
	ttl time.Duration
}

// NewConfig contructs a Config.
// Cache expires ttl+rand.Int63n(jitter) after the value has been generated.
func NewConfig(ttl time.Duration, jitter time.Duration) *Config {
	if ttl != NoExpiration && jitter > 0 {
		ttl += time.Duration(r.Int63n(int64(jitter)))
	}
	return &Config{ttl}
}

// NoExpiration is a constant passed to NewConfig when you do not want to expire the cache.
const NoExpiration time.Duration = -1

// DefaultConfig is 1 hour for ttl and 5 minutes for jitter.
func DefaultConfig() *Config {
	return NewConfig(time.Hour, time.Minute*5)
}

func (c *Config) expiresAt() time.Time {
	if c.ttl == NoExpiration {
		// https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go
		return time.Unix(1<<63-62135596801, 999999999)
	}
	return time.Now().Add(c.ttl)
}

// Get retrieves the value stored in the cache.
// If the value is invalid, it calls an anonymous function given as an argument and stores it in the cache.
func (c *Cache) Get(initValue func() (interface{}, error)) (interface{}, error) {
	if c.expiresAt.After(time.Now()) {
		return c.value, nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.expiresAt.After(time.Now()) {
		return c.value, nil
	}
	value, err := initValue()
	if err != nil {
		return nil, err
	}
	c.value = value
	c.expiresAt = c.config.expiresAt()
	return value, nil
}

// Invalidate will invalidate the cache.
func (c *Cache) Invalidate() {
	c.lock.Lock()
	c.expiresAt = time.Now()
	c.lock.Unlock()
}
