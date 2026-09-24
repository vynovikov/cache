package ttl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ttlSuite struct {
	suite.Suite
}

func TestTTLSuite(t *testing.T) {
	suite.Run(t, new(ttlSuite))
}
func (s *ttlSuite) TestShiftUp() {
	tt := []struct {
		name         string
		cap          int
		TTLNodes     []*Node
		currentIndex int
		wantTTLNodes []*Node
	}{
		{
			name: "0. Moved to intermediate level",
			cap:  10,
			TTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(5 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 6},
				{Key: "08", ExpireAt: time.Now().Add(7 * time.Second), HeapIndex: 7},
			},
			currentIndex: 7,
			wantTTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(5 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "08", ExpireAt: time.Now().Add(7 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 6},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 7},
			},
		},
		{
			name: "1. Moved to top level",
			cap:  10,
			TTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(5 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 6},
				{Key: "08", ExpireAt: time.Now().Add(4 * time.Second), HeapIndex: 7},
			},
			currentIndex: 7,
			wantTTLNodes: []*Node{
				{Key: "08", ExpireAt: time.Now().Add(4 * time.Second), HeapIndex: 0},
				{Key: "01", ExpireAt: time.Now().Add(5 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 6},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 7},
			},
		},
		{
			name: "2. Not moved",
			cap:  10,
			TTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(5 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 6},
				{Key: "08", ExpireAt: time.Now().Add(13 * time.Second), HeapIndex: 7},
			},
			currentIndex: 7,
			wantTTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(5 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 6},
				{Key: "08", ExpireAt: time.Now().Add(13 * time.Second), HeapIndex: 7},
			},
		},
		{
			name: "3. Real 1",
			cap:  10,
			TTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(80 * time.Millisecond), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(70 * time.Millisecond), HeapIndex: 1},
			},
			currentIndex: 1,
			wantTTLNodes: []*Node{
				{Key: "02", ExpireAt: time.Now().Add(70 * time.Millisecond), HeapIndex: 0},
				{Key: "01", ExpireAt: time.Now().Add(80 * time.Millisecond), HeapIndex: 1},
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			// 0. Initialize heap
			heap := Heap{
				Nodes: v.TTLNodes,
			}

			// 1. ShiftUp
			heap.ShiftUp(v.currentIndex)

			// 2. Comparing results
			for j, w := range heap.Nodes {
				s.Equal(v.wantTTLNodes[j].HeapIndex, w.HeapIndex)
				s.WithinDuration(w.ExpireAt, v.wantTTLNodes[j].ExpireAt, 5*time.Millisecond)
				s.Equal(v.wantTTLNodes[j].Key, w.Key)
			}
		})
	}
}

func (s *ttlSuite) TestShiftDown() {
	tt := []struct {
		name         string
		cap          int
		TTLNodes     []*Node
		currentIndex int
		wantTTLNodes []*Node
	}{
		{
			name: "0. Moved to intermediate level",
			TTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
			currentIndex: 0,
			wantTTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
		},
		{
			name: "1. Moved to bottom level",
			TTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(16 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
			currentIndex: 0,
			wantTTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(16 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
		},
		{
			name: "2. Not moved",
			TTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(4 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
			currentIndex: 0,
			wantTTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(4 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			// 0. Initialize heap
			heap := Heap{
				Nodes: v.TTLNodes,
			}

			// 1. ShiftDown
			heap.ShiftDown(v.currentIndex)

			// 2 Comparing results
			for j, w := range heap.Nodes {
				s.Equal(v.wantTTLNodes[j].HeapIndex, w.HeapIndex)
				s.WithinDuration(w.ExpireAt, v.wantTTLNodes[j].ExpireAt, 5*time.Millisecond)
			}
		})
	}
}

func (s *ttlSuite) TestRebalance() {
	tt := []struct {
		name         string
		cap          int
		TTLNodes     []*Node
		wantTTLNodes []*Node
	}{
		{
			name: "0. Top is unbalanced",
			cap:  10,
			TTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
			wantTTLNodes: []*Node{
				{Key: "02", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 0},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "01", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
		},
		{
			name: "1. Intermediate is unbalanced",
			cap:  10,
			TTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
			wantTTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 0},
				{Key: "04", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "02", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
		},
		{
			name: "2. Balanced",
			cap:  10,
			TTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
			wantTTLNodes: []*Node{
				{Key: "01", ExpireAt: time.Now().Add(6 * time.Second), HeapIndex: 0},
				{Key: "02", ExpireAt: time.Now().Add(9 * time.Second), HeapIndex: 1},
				{Key: "03", ExpireAt: time.Now().Add(8 * time.Second), HeapIndex: 2},
				{Key: "04", ExpireAt: time.Now().Add(10 * time.Second), HeapIndex: 3},
				{Key: "05", ExpireAt: time.Now().Add(11 * time.Second), HeapIndex: 4},
				{Key: "06", ExpireAt: time.Now().Add(12 * time.Second), HeapIndex: 5},
				{Key: "07", ExpireAt: time.Now().Add(14 * time.Second), HeapIndex: 6},
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			// 0. Initialize heap
			heap := Heap{
				Nodes: v.TTLNodes,
			}

			// 0.1 Waiting
			time.Sleep(time.Millisecond * 20)

			// 1 Rebalance
			heap.Rebalance()

			// 2 Comparing results
			for j, w := range heap.Nodes {
				s.Equal(v.wantTTLNodes[j].HeapIndex, w.HeapIndex)
				s.Equal(v.wantTTLNodes[j].Key, w.Key)
				s.WithinDuration(w.ExpireAt, v.wantTTLNodes[j].ExpireAt, 5*time.Millisecond)
			}
		})
	}
}
