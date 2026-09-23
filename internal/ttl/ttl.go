package ttl

import "time"

type Node struct {
	Key       string
	ExpireAt  time.Time
	HeapIndex int
}

type Heap struct {
	Nodes []*Node
}

func NewHeap(cap int) Heap {
	Nodes := make([]*Node, 0, cap)

	return Heap{
		Nodes: Nodes,
	}
}

func (h *Heap) ShiftUp(currentIndex int) {
	for currentIndex > 0 {

		parentIndex := (currentIndex - 1) / 2
		if h.Nodes[currentIndex].ExpireAt.After(h.Nodes[parentIndex].ExpireAt) {
			break
		}

		h.Swap(currentIndex, parentIndex)

		currentIndex = parentIndex
	}
}

func (h *Heap) ShiftDown(currentIndex int) {
	nodesLen := len(h.Nodes)
	for {
		leftChildIndex := currentIndex*2 + 1
		rightChildIndex := currentIndex*2 + 2
		smallestIndex := currentIndex

		if leftChildIndex < nodesLen && h.Nodes[leftChildIndex].ExpireAt.Before(h.Nodes[smallestIndex].ExpireAt) {
			smallestIndex = leftChildIndex
		}

		if rightChildIndex < nodesLen && h.Nodes[rightChildIndex].ExpireAt.Before(h.Nodes[smallestIndex].ExpireAt) {
			smallestIndex = rightChildIndex
		}

		if smallestIndex != currentIndex {

			h.Swap(currentIndex, smallestIndex)

			currentIndex = smallestIndex

			continue
		}

		break
	}
}

func (h *Heap) Swap(i, j int) {
	h.Nodes[i], h.Nodes[j] = h.Nodes[j], h.Nodes[i]

	h.Nodes[i].HeapIndex = i
	h.Nodes[j].HeapIndex = j
}

func (h *Heap) Remove(removeIndex int) {
	heapLastIndex := len(h.Nodes) - 1

	if removeIndex == heapLastIndex {
		h.Nodes = h.Nodes[:removeIndex]

	} else {
		h.Swap(removeIndex, heapLastIndex)
		h.Nodes = h.Nodes[:heapLastIndex]
		h.ShiftUp(removeIndex)
		h.ShiftDown(removeIndex)
	}
}

func (h *Heap) Rebalance() {
	n := len(h.Nodes)

	for i := n/2 - 1; i >= 0; i-- {
		h.ShiftDown(i)
	}
}

func (h *Heap) RemoveHeapExpired(now time.Time) []string {
	expiredCount := 0
	for expiredCount < len(h.Nodes) {
		if h.Nodes[expiredCount].ExpireAt.After(now) {
			break
		}
		expiredCount++
	}

	if expiredCount == 0 {
		return nil
	}

	expiredKeys := make([]string, expiredCount)
	for i := 0; i < expiredCount; i++ {
		expiredKeys[i] = h.Nodes[i].Key
		h.Nodes[i] = nil
	}

	h.Nodes = h.Nodes[expiredCount:]

	for i, node := range h.Nodes {
		node.HeapIndex = i
	}

	return expiredKeys
}
