package cache

import (
	"sync"
	"time"
)

type TTLLRUCache interface {
	Set(key string, value any, ttl time.Duration)
	Get(key string) (value any, exists bool)

	// Freeze stops the background eviction worker to conserve CPU resources.
	// The cache remains fully accessible for concurrent reads (Get) and writes (Set).
	// After calling Freeze, expired items are evicted lazily upon calling Get.
	Freeze()
}

type LRUNode struct {
	Key  string
	Next *LRUNode
	Prev *LRUNode
}

type TTLNode struct {
	Key       string
	expireAt  time.Time
	heapIndex int
}

type CacheNode struct {
	Value   any
	LRUElem *LRUNode
	TTLElem *TTLNode
}

func newCacheNode(key string, value any, expireAt time.Time, TTLHeapIndex int) *CacheNode {

	return &CacheNode{
		Value: value,
		LRUElem: &LRUNode{
			Key: key,
		},
		TTLElem: &TTLNode{
			Key:       key,
			expireAt:  expireAt,
			heapIndex: TTLHeapIndex,
		},
	}
}

func newTTLNodeFromCacheNode(cacheNode *CacheNode) *TTLNode {

	return &TTLNode{
		Key:       cacheNode.TTLElem.Key,
		heapIndex: cacheNode.TTLElem.heapIndex,
		expireAt:  cacheNode.TTLElem.expireAt,
	}
}

type LRULinkedList struct {
	head *LRUNode
	tail *LRUNode
}

func newLRULinkedList() LRULinkedList {
	head, tail := &LRUNode{}, &LRUNode{}
	head.Next = tail
	tail.Prev = head

	return LRULinkedList{
		head: head,
		tail: tail,
	}
}

type TTLHeap struct {
	TTLNodes []*TTLNode
}

func newTTLHeap(cap int) TTLHeap {
	TTLNodes := make([]*TTLNode, 0, cap)

	return TTLHeap{
		TTLNodes: TTLNodes,
	}
}

type TTLLRUCacheImpl struct {
	mu    sync.Mutex
	data  map[string]*CacheNode
	cap   int
	LRULL LRULinkedList
	TTLH  TTLHeap

	doneChan       chan struct{}
	freezeChan     chan struct{}
	resetTimerChan chan struct{}
}

func NewCache(cap int) *TTLLRUCacheImpl {
	m := make(map[string]*CacheNode, cap)
	ttlHeap := newTTLHeap(cap)
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
		oldCacheNode.TTLElem.expireAt = time.Now().Add(ttl)

		c.data[key] = oldCacheNode

		c.TTLH.shiftUp(oldCacheNode.TTLElem.heapIndex)
		c.TTLH.shiftDown(oldCacheNode.TTLElem.heapIndex)
		c.LRULL.moveToHead(oldCacheNode.LRUElem)

		if oldCacheNode.TTLElem.heapIndex == 0 {

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
		keyToDelete := c.LRULL.removeTail()
		heapIndexToRemove := c.data[keyToDelete].TTLElem.heapIndex
		c.TTLH.remove(heapIndexToRemove)

		delete(c.data, keyToDelete)
	}

	newCacheNodeExample := newCacheNode(key, value, time.Now().Add(ttl), len(c.TTLH.TTLNodes))

	c.data[key] = newCacheNodeExample
	c.TTLH.TTLNodes = append(c.TTLH.TTLNodes, newCacheNodeExample.TTLElem)
	c.LRULL.insertAtHead(newCacheNodeExample.LRUElem)

	heapCurrentSize := len(c.TTLH.TTLNodes)
	if key == "key01" {
		x := 1
		_ = x
	}
	c.TTLH.shiftUp(heapCurrentSize - 1)

	if newCacheNodeExample.TTLElem.heapIndex == 0 && len(c.TTLH.TTLNodes) > 1 {

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

		if foundCacheNode.TTLElem.expireAt.Before(time.Now()) {
			c.LRULL.remove(foundCacheNode.LRUElem)
			heapIndexToRemove := c.data[key].TTLElem.heapIndex
			c.TTLH.remove(heapIndexToRemove)

			delete(c.data, key)

			return nil, false
		}

		c.LRULL.moveToHead(foundCacheNode.LRUElem)

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

	expiredKeys := c.TTLH.removeHeapExpired(time.Now())

	for _, key := range expiredKeys {
		if cacheNode, ok := c.data[key]; ok {
			c.LRULL.remove(cacheNode.LRUElem)
			delete(c.data, key)
		}
	}

	if len(c.TTLH.TTLNodes) > 0 {
		timeLeft := max(time.Until(c.TTLH.TTLNodes[0].expireAt), 0)
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

func (l *LRULinkedList) remove(LRUnode *LRUNode) {
	LRUnode.Prev.Next = LRUnode.Next
	LRUnode.Next.Prev = LRUnode.Prev
}

func (l *LRULinkedList) insertAtHead(LRUnode *LRUNode) {
	LRUnode.Prev = l.head
	LRUnode.Next = l.head.Next

	l.head.Next.Prev = LRUnode
	l.head.Next = LRUnode
}

func (l *LRULinkedList) moveToHead(LRUnode *LRUNode) {
	l.remove(LRUnode)
	l.insertAtHead(LRUnode)
}

func (l *LRULinkedList) removeTail() string {
	tailPrevKey := l.tail.Prev.Key

	l.remove(l.tail.Prev)

	return tailPrevKey
}

func (h *TTLHeap) shiftUp(currentIndex int) {
	for currentIndex > 0 {

		parentIndex := (currentIndex - 1) / 2
		if h.TTLNodes[currentIndex].expireAt.After(h.TTLNodes[parentIndex].expireAt) {
			break
		}

		h.swap(currentIndex, parentIndex)

		currentIndex = parentIndex
	}
}

func (h *TTLHeap) shiftDown(currentIndex int) {
	nodesLen := len(h.TTLNodes)
	for {
		leftChildIndex := currentIndex*2 + 1
		rightChildIndex := currentIndex*2 + 2
		smallestIndex := currentIndex

		if leftChildIndex < nodesLen && h.TTLNodes[leftChildIndex].expireAt.Before(h.TTLNodes[smallestIndex].expireAt) {
			smallestIndex = leftChildIndex
		}

		if rightChildIndex < nodesLen && h.TTLNodes[rightChildIndex].expireAt.Before(h.TTLNodes[smallestIndex].expireAt) {
			smallestIndex = rightChildIndex
		}

		if smallestIndex != currentIndex {

			h.swap(currentIndex, smallestIndex)

			currentIndex = smallestIndex

			continue
		}

		break
	}
}

func (h *TTLHeap) swap(i, j int) {
	h.TTLNodes[i], h.TTLNodes[j] = h.TTLNodes[j], h.TTLNodes[i]

	h.TTLNodes[i].heapIndex = i
	h.TTLNodes[j].heapIndex = j
}

func (h *TTLHeap) remove(removeIndex int) {
	heapLastIndex := len(h.TTLNodes) - 1

	if removeIndex == heapLastIndex {
		h.TTLNodes = h.TTLNodes[:removeIndex]

	} else {
		h.swap(removeIndex, heapLastIndex)
		h.TTLNodes = h.TTLNodes[:heapLastIndex]
		h.shiftUp(removeIndex)
		h.shiftDown(removeIndex)
	}
}

func (h *TTLHeap) rebalance() {
	n := len(h.TTLNodes)

	for i := n/2 - 1; i >= 0; i-- {
		h.shiftDown(i)
	}
}

func (h *TTLHeap) removeHeapExpired(now time.Time) []string {
	expiredCount := 0
	for expiredCount < len(h.TTLNodes) {
		if h.TTLNodes[expiredCount].expireAt.After(now) {
			break
		}
		expiredCount++
	}

	if expiredCount == 0 {
		return nil
	}

	expiredKeys := make([]string, expiredCount)
	for i := 0; i < expiredCount; i++ {
		expiredKeys[i] = h.TTLNodes[i].Key
		h.TTLNodes[i] = nil
	}

	h.TTLNodes = h.TTLNodes[expiredCount:]

	for i, node := range h.TTLNodes {
		node.heapIndex = i
	}

	return expiredKeys
}
