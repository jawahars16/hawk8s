package cache

import "time"

type cache struct {
	data     map[string]cacheValue
	validity int
}

type cacheValue struct {
	data    interface{}
	created int64
}

func New(validityInSeconds int) *cache {
	return &cache{
		data:     make(map[string]cacheValue),
		validity: validityInSeconds,
	}
}

func (c *cache) Get(key string) (interface{}, bool) {
	val, ok := c.data[key]
	if val.created+int64(c.validity) < time.Now().Unix() {
		delete(c.data, key)
		return nil, false
	}
	return val.data, ok
}

func (c *cache) Set(key string, value interface{}) {
	c.data[key] = cacheValue{
		data:    value,
		created: time.Now().Unix(),
	}
}
