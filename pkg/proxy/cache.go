package proxy

import (
	"net/http"
	"sync"
)

// CacheEntry represents a cached HTTP response
type CacheEntry struct {
	Status  int
	Headers http.Header
	Body    []byte
	Size    int // bytes
}

// Doubly-linked list node
type lruNode struct {
	key  string
	val  *CacheEntry
	prev *lruNode
	next *lruNode
}

// LRUCache is an in-memory least-recently-used cache
type LRUCache struct {
	mu       sync.Mutex
	capBytes int
	used     int

	items map[string]*lruNode
	head  *lruNode // most recently used
	tail  *lruNode // least recently used
}

// NewLRUCache creates a new LRU cache with max capacity in bytes.
// If capBytes <= 0, caching is disabled (returns nil).
func NewLRUCache(capBytes int) *LRUCache {
	if capBytes <= 0 {
		return nil
	}
	return &LRUCache{
		capBytes: capBytes,
		items:    make(map[string]*lruNode),
	}
}

// Get retrieves a cache entry and marks it as most recently used.
func (c *LRUCache) Get(key string) (*CacheEntry, bool) {
	if c == nil {
		return nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.moveToFront(node)
	return node.val, true
}

// Put inserts or updates a cache entry.
func (c *LRUCache) Put(key string, entry *CacheEntry) {
	if c == nil || entry == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing
	if node, ok := c.items[key]; ok {
		c.used -= node.val.Size
		node.val = entry
		c.used += entry.Size
		c.moveToFront(node)
	} else {
		node := &lruNode{
			key: key,
			val: entry,
		}
		c.items[key] = node
		c.addToFront(node)
		c.used += entry.Size
	}

	// Evict until under capacity
	for c.used > c.capBytes {
		c.evict()
	}
}

// ─────────────────────────────────────────────
// Internal helpers

func (c *LRUCache) addToFront(n *lruNode) {
	n.prev = nil
	n.next = c.head

	if c.head != nil {
		c.head.prev = n
	}
	c.head = n

	if c.tail == nil {
		c.tail = n
	}
}

func (c *LRUCache) moveToFront(n *lruNode) {
	if c.head == n {
		return
	}

	// detach
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	if c.tail == n {
		c.tail = n.prev
	}

	// move to front
	n.prev = nil
	n.next = c.head
	if c.head != nil {
		c.head.prev = n
	}
	c.head = n

	if c.tail == nil {
		c.tail = n
	}
}

func (c *LRUCache) evict() {
	if c.tail == nil {
		return
	}

	node := c.tail

	if node.prev != nil {
		node.prev.next = nil
	}
	c.tail = node.prev

	if c.tail == nil {
		c.head = nil
	}

	delete(c.items, node.key)
	c.used -= node.val.Size
}
