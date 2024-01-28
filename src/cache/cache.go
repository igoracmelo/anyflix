package cache

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	m      *sync.Map
	ttls   map[K]time.Time
	ttlsMu *sync.Mutex
}

func New[K comparable, V any]() Cache[K, V] {
	return Cache[K, V]{
		m:      &sync.Map{},
		ttls:   map[K]time.Time{},
		ttlsMu: &sync.Mutex{},
	}
}

func (c Cache[K, V]) Get(key K) (V, bool) {
	v, ok := c.m.Load(key)
	var val V
	if v != nil {
		val = v.(V)
	}
	return val, ok
}

func (c Cache[K, V]) Set(key K, val V, dur time.Duration) {
	c.m.Swap(key, val)
	if dur == 0 {
		return
	}

	go func() {
		c.ttlsMu.Lock()
		c.ttls[key] = time.Now().Add(dur)
		c.ttlsMu.Unlock()

		time.Sleep(dur)

		c.ttlsMu.Lock()
		defer c.ttlsMu.Unlock()

		currTTL := c.ttls[key]
		if currTTL != (time.Time{}) && time.Now().After(currTTL) {
			c.Delete(key)
			c.ttls[key] = time.Time{}
		}
	}()
}

func (c Cache[K, V]) Delete(key K) {
	c.m.Delete(key)
}

var ErrCacheMiss = errors.New("cache miss")

type FileCache struct {
	mu  *sync.Mutex
	dir string
	ttl time.Duration
}

func NewFileCache(baseDir string, ttl time.Duration) FileCache {
	return FileCache{
		mu:  &sync.Mutex{},
		dir: baseDir,
		ttl: ttl,
	}
}

func (c FileCache) Get(key string) (b []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sum := md5.Sum([]byte(key))
	path := filepath.Join(c.dir, hex.EncodeToString(sum[:]))

	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		err = ErrCacheMiss
		return
	}
	if err != nil {
		log.Print(err)
		return
	}

	if c.ttl != 0 && time.Since(info.ModTime()) > c.ttl {
		err = os.Remove(path)
		if err != nil {
			log.Print(err)
			return
		}
		err = ErrCacheMiss
		return
	}

	slog.Info("reading from cache", "path", path)
	b, err = os.ReadFile(path)
	return
}

func (c FileCache) Set(key string, val []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sum := md5.Sum([]byte(key))
	path := filepath.Join(c.dir, hex.EncodeToString(sum[:]))

	return os.WriteFile(path, val, 0666)
}

func (c FileCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sum := md5.Sum([]byte(key))
	path := filepath.Join(c.dir, hex.EncodeToString(sum[:]))
	return os.Remove(path)
}

func (c FileCache) TTL() time.Duration {
	return c.ttl
}
