package inmem_cache

import (
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type InmemCache[T any] struct {
	cache *ttlcache.Cache[string, T]
}

func New[T any]() *InmemCache[T] {
	c := &InmemCache[T]{}
	c.cache = ttlcache.New(ttlcache.WithDisableTouchOnHit[string, T]())
	return c
}

func (c InmemCache[T]) Set(key string, value T, ttlSeconds ...int) error {

	ttl := ttlcache.NoTTL
	if len(ttlSeconds) > 0 {
		ttl = time.Second * time.Duration(ttlSeconds[0])
	}

	c.cache.Set(key, value, ttl)

	return nil
}

func (c InmemCache[T]) Get(key string, value *T) (bool, error) {

	item := c.cache.Get(key)
	if item == nil || item.IsExpired() {
		return false, nil
	}

	*value = item.Value()
	return true, nil
}

func (c InmemCache[T]) Unset(key string) error {

	c.cache.Delete(key)

	return nil
}

func (c InmemCache[T]) Clear() error {

	c.cache.DeleteAll()

	return nil
}

func (c InmemCache[T]) Touch(key string) error {

	c.cache.Touch(key)

	return nil
}

func (c InmemCache[T]) Start() {
	go c.cache.Start()
}

func (c InmemCache[T]) Stop() {
	c.cache.Stop()
}

func (c InmemCache[T]) Keys() ([]string, error) {

	keys := c.cache.Keys()

	return keys, nil
}
