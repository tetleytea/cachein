package lru

import (
	lst "container/list"
	"errors"
	"sync"
	"time"
)

type LRUCacheNode struct {
	key   string
	value interface{}
	ts    int64
}

type LRUCache struct {
	maxSize     int64
	currentSize int64
	l           *lst.List
	cache       map[string]*lst.Element
	mutex       *sync.Mutex
	epxiry      time.Duration
}

func NewCache(maxSize int64) (*LRUCache, error) {
	if maxSize <= 0 {
		return nil, errors.New("LRUCache maxSize should be larger than 0")
	}

	return &LRUCache{
		maxSize: maxSize,
		cache:   make(map[string]*lst.Element),
		l:       lst.New(),
		mutex: 	 &sync.Mutex{},
	}, nil
}

func (c *LRUCache) Remove(key string) {
	c.mutex.Lock()
	if entry, hit := c.cache[key]; hit {
		c.RemoveEntry(entry)
	}
	c.mutex.Unlock()
}

func (c *LRUCache) Add(key string, value interface{}) {
	c.mutex.Lock()
	var ts int64
	if c.epxiry != time.Duration(0) {
		ts = time.Now().UnixNano() / int64(time.Millisecond)
	}
	if entry, ok := c.cache[key]; ok {
		c.l.MoveToFront(entry)
		entry.Value.(*LRUCacheNode).value = value
		entry.Value.(*LRUCacheNode).ts = ts
		return
	}
	ele := c.l.PushFront(&LRUCacheNode{key, value, ts})
	c.cache[key] = ele
	if c.maxSize != 0 && c.Size() > c.maxSize {
		c.RemoveOldest()
	}
	c.mutex.Unlock()
}

func (c *LRUCache) Get(key string) (value interface{}, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entry, hit := c.cache[key]; hit {
		c.l.MoveToFront(entry)
		return entry.Value.(*LRUCacheNode).value, true
	}
	return
}

func (c *LRUCache) Size() int64 {
	return int64(c.l.Len())
}

func (c *LRUCache) RemoveEntry(e *lst.Element) {
	c.l.Remove(e)
	kv := e.Value.(*LRUCacheNode)
	delete(c.cache, kv.key)
}

func (c *LRUCache) RemoveOldest() {
	entry := c.l.Back()
	if entry != nil {
		c.RemoveEntry(entry)
	}
}