package cache

import (
	"sync"
	"time"

	"github.com/cache/main/internal/lru"
	"github.com/cache/main/internal/ttl"
)

type TTLLRUCache interface {
	Set(key string, value any, ttl time.Duration)
	Get(key string) (value any, exists bool)

	// Freeze stops the background eviction worker to conserve CPU resources.
	// The cache remains fully accessible for concurrent reads (Get) and writes (Set).
	// After calling Freeze, expired items are evicted lazily upon calling Get.
	Freeze()
}

type CacheNode struct {
	Value   any
	LRUElem *lru.Node
	TTLElem *ttl.Node
}

func newCacheNode(key string, value any, expireAt time.Time, TTLHeapIndex int) *CacheNode {

	return &CacheNode{
		Value: value,
		LRUElem: &lru.Node{
			Key: key,
		},
		TTLElem: &ttl.Node{
			Key:       key,
			ExpireAt:  expireAt,
			HeapIndex: TTLHeapIndex,
		},
	}
}

func newTTLNodeFromCacheNode(cacheNode *CacheNode) *ttl.Node {

	return &ttl.Node{
		Key:       cacheNode.TTLElem.Key,
		HeapIndex: cacheNode.TTLElem.HeapIndex,
		ExpireAt:  cacheNode.TTLElem.ExpireAt,
	}
}

func newLRULinkedList() lru.LinkedList {
	head, tail := &lru.Node{}, &lru.Node{}
	head.Next = tail
	tail.Prev = head

	return lru.LinkedList{
		Head: head,
		Tail: tail,
	}
}

type TTLLRUCacheImpl struct {
	mu    sync.Mutex
	data  map[string]*CacheNode
	cap   int
	LRULL lru.LinkedList
	TTLH  ttl.Heap

	doneChan       chan struct{}
	freezeChan     chan struct{}
	resetTimerChan chan struct{}
}

func NewCache(cap int) *TTLLRUCacheImpl {
	m := make(map[string]*CacheNode, cap)
	ttlHeap := ttl.NewHeap(cap)
	lruLL := newLRULinkedList()

	doneChan := make(chan struct{}, 1)
	freezeChan := make(chan struct{})
	reserTimerChan := make(chan struct{}, 1)

	cache := &TTLLRUCacheImpl{
		mu:    sync.Mutex{},
		data:  m,
		cap:   cap,
		LRULL: lruLL,
		TTLH:  ttlHeap,

		doneChan:       doneChan,
		freezeChan:     freezeChan,
		resetTimerChan: reserTimerChan,
	}

	go cache.startCleanupWorker()

	return cache
}

func (c *TTLLRUCacheImpl) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if oldCacheNode, ok := c.data[key]; ok {
		oldCacheNode.Value = value
		oldCacheNode.TTLElem.ExpireAt = time.Now().Add(ttl)

		c.data[key] = oldCacheNode

		c.TTLH.ShiftUp(oldCacheNode.TTLElem.HeapIndex)
		c.TTLH.ShiftDown(oldCacheNode.TTLElem.HeapIndex)
		c.LRULL.MoveToHead(oldCacheNode.LRUElem)

		if oldCacheNode.TTLElem.HeapIndex == 0 {

			select {
			case c.resetTimerChan <- struct{}{}:
			default:
				// If the channel buffer is full, drop the signal and proceed.
				// The worker is guaranteed to wake up since a signal is already pending.
			}
		}

		return
	}

	// Not found

	if len(c.data) == c.cap {
		keyToDelete := c.LRULL.RemoveTail()
		heapIndexToRemove := c.data[keyToDelete].TTLElem.HeapIndex
		c.TTLH.Remove(heapIndexToRemove)

		delete(c.data, keyToDelete)
	}

	newCacheNodeExample := newCacheNode(key, value, time.Now().Add(ttl), len(c.TTLH.Nodes))

	c.data[key] = newCacheNodeExample
	c.TTLH.Nodes = append(c.TTLH.Nodes, newCacheNodeExample.TTLElem)
	c.LRULL.InsertAtHead(newCacheNodeExample.LRUElem)

	heapCurrentSize := len(c.TTLH.Nodes)
	if key == "key01" {
		x := 1
		_ = x
	}
	c.TTLH.ShiftUp(heapCurrentSize - 1)

	if newCacheNodeExample.TTLElem.HeapIndex == 0 && len(c.TTLH.Nodes) > 1 {

		select {
		case c.resetTimerChan <- struct{}{}:
		default:
			// If the channel buffer is full, drop the signal and proceed.
			// The worker is guaranteed to wake up since a signal is already pending.
		}
	}
}

func (c *TTLLRUCacheImpl) Get(key string) (value any, exists bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if foundCacheNode, ok := c.data[key]; ok {

		if foundCacheNode.TTLElem.ExpireAt.Before(time.Now()) {
			c.LRULL.Remove(foundCacheNode.LRUElem)
			heapIndexToRemove := c.data[key].TTLElem.HeapIndex
			c.TTLH.Remove(heapIndexToRemove)

			delete(c.data, key)

			return nil, false
		}

		c.LRULL.MoveToHead(foundCacheNode.LRUElem)

		return foundCacheNode.Value, true
	}

	// not found

	return nil, false
}

func (c *TTLLRUCacheImpl) Freeze() {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case <-c.freezeChan:
		return
	default:
	}

	close(c.freezeChan)

	<-c.doneChan
}

func (c *TTLLRUCacheImpl) removeCacheExpired() (time.Duration, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiredKeys := c.TTLH.RemoveHeapExpired(time.Now())

	for _, key := range expiredKeys {
		if cacheNode, ok := c.data[key]; ok {
			c.LRULL.Remove(cacheNode.LRUElem)
			delete(c.data, key)
		}
	}

	if len(c.TTLH.Nodes) > 0 {
		timeLeft := max(time.Until(c.TTLH.Nodes[0].ExpireAt), 0)
		return timeLeft, true
	}

	return 0, false
}

func (c *TTLLRUCacheImpl) startCleanupWorker() {
	// Гарантируем, что при выходе из горутины Close() получит сигнал о завершении
	defer close(c.doneChan)

	// Создаем изначально остановленный таймер, чтобы не тикал вхолостую
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}

	for {
		timeLeft, hasElements := c.removeCacheExpired()

		if hasElements {
			timer.Reset(timeLeft)
		} else {
			timer.Stop()
		}

		select {
		case <-c.freezeChan:
			return

		case <-timer.C:

		case <-c.resetTimerChan:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		}
	}
}
