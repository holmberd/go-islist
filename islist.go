// Package islist implements an in-memory Interval Skiplist probabilistic data structure.
// It supports efficient insertion, deletion, and querying for intervals that overlap a given range.
//
// Time and space complexity, see: https://en.wikipedia.org/wiki/Skip_list
package islist

import (
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"os"
)

const (
	MaxLevel               = 32   // log_(4)(2^64) = 32.
	MaxSearchLevel         = 10   // log_(4)(10^6) = ~10
	Probability    float32 = 0.25 // P = 1/4
	levelThreshold int32   = int32(Probability * math.MaxInt32)
)

// SkipList represent an Interval Skiplist probabilistic data structure for contiguous intervals.
//
// A Skiplist assigns levels to nodes randomly using a geometric distribution.
// Each level (k) has nodes with probability (p^k), where (p) is the probability factor.
// The theoretical maximum level (L) of a skiplist grows logarithmically with the number
// of elements (n): L = log_(1/p)(n)
type SkipList struct {
	head     *Node
	maxLevel int
	length   int
	pool     *NodePool
	PCG      *rand.PCG
}

// New returns a new instance of a SkipList.
func New(pool *NodePool, PCG *rand.PCG) *SkipList {
	return &SkipList{
		head:     newNode(pool, MaxLevel, IntervalKey{}),
		maxLevel: 1,
		length:   0,
		pool:     pool,
		PCG:      PCG,
	}
}

// QueryParam represent parameters used in list queries.
type QueryParam struct {
	Offset int
	Limit  int
}

// newNode returns a new instance of a node.
func newNode(pool *NodePool, level int, ik IntervalKey) *Node {
	n := pool.get()
	if cap(n.levels) >= level {
		// Reuse any preallocated capacity.
		n.levels = n.levels[:level]
	} else {
		n.levels = make([]nodeLevel, level)
	}
	n.intervalKey = ik
	return n
}

// randomLevel returns a random level.
func (sl *SkipList) randomLevel() int {
	r := rand.New(sl.PCG)
	level := 1
	for r.Int32() < levelThreshold && level < MaxLevel {
		level++
	}
	return level
}

// Insert adds a new key to the list.
// If the key already exist, it updates the existing key and returns the previous key.
func (sl *SkipList) Insert(intervalKey IntervalKey) *IntervalKey {
	var n *Node
	var i int
	nodePath := make([]*Node, MaxLevel) // Top-to-bottom path to the inserted node.
	dist := make([]int, MaxLevel)       // Tracks the cumulative distance (span) traveled at each level.

	// Find the position to insert the new node, top level down search.
	n = sl.head
	for i = sl.maxLevel - 1; i >= 0; i-- {
		if i < len(dist)-1 {
			dist[i] = dist[i+1] // Initialize with travelled distance from the level above.
		}
		// Positions n at the last node whose interval does not exceed the new interval's start.
		for n.levels[i].next != nil && less(n.levels[i].next.intervalKey, intervalKey) {
			dist[i] += n.levels[i].span // Accumulate span traversed.
			n = n.levels[i].next
		}
		nodePath[i] = n // Populate for each level.
	}

	if n.levels[0].next != nil && n.levels[0].next.intervalKey.equalInterval(intervalKey) {
		// Interval exists. Update the node's key.
		xn := n.levels[0].next
		xk := xn.intervalKey
		xn.intervalKey = intervalKey
		return &xk
	}

	// Create a new node for the new key and link it.
	rLevel := sl.randomLevel()
	n = newNode(sl.pool, rLevel, intervalKey)
	for i, insertMaxLevel := 0, max(sl.maxLevel, rLevel); i < insertMaxLevel; i++ {
		if i >= sl.maxLevel {
			// Initialize any new higher levels.
			nodePath[i] = sl.head
			nodePath[i].levels[i].span = sl.length
			sl.maxLevel++
		}
		if i < rLevel {
			// Link the new node(n) up to the random level: n1 -> n2 => n1 -> n -> n2
			n.levels[i].next = nodePath[i].levels[i].next // n -> n2
			n.levels[i].span = nodePath[i].levels[i].span - (dist[0] - dist[i])
			nodePath[i].levels[i].next = n // n1 -> n
			nodePath[i].levels[i].span = (dist[0] - dist[i]) + 1
		} else {
			// Adjust spans for any levels above the random level.
			nodePath[i].levels[i].span++
		}
	}
	sl.length++
	return nil
}

// Delete removes a key with the specified interval.
// Returns the key of the deleted node if found.
func (sl *SkipList) Delete(interval IntervalKey) *IntervalKey {
	var k IntervalKey
	var n *Node
	var i int
	nodePath := make([]*Node, MaxLevel)

	// Find the node to delete.
	n = sl.head
	for i = sl.maxLevel - 1; i >= 0; i-- {
		for n.levels[i].next != nil && less(n.levels[i].next.intervalKey, interval) {
			n = n.levels[i].next
		}
		nodePath[i] = n
	}
	n = n.levels[0].next
	if n == nil || !n.intervalKey.equalInterval(interval) {
		return nil
	}

	// Delete/Unlink the node.
	ml := sl.maxLevel
	for i := 0; i < ml; i++ {
		// Levels where the node exists.
		if i < len(n.levels) && nodePath[i].levels[i].next == n {
			nodePath[i].levels[i].next = n.levels[i].next
			nodePath[i].levels[i].span += n.levels[i].span - 1
			if (sl.maxLevel > i && sl.maxLevel > 1) && sl.head.levels[i].next == nil {
				sl.maxLevel = i // Adjust maxLevel to the highest level that contain nodes.
			}
		} else {
			// Levels beyond the node's levels.
			nodePath[i].levels[i].span--
		}
	}
	k = n.intervalKey
	sl.pool.put(n)
	sl.length--
	return &k
}

// Overlaps returns all keys that overlap the query interval.
func (sl *SkipList) Overlaps(interval IntervalKey, qParam QueryParam) (result []*IntervalKey) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println()
			sl.Print(os.Stdout, MaxLevel)
			// fmt.Printf("\nlevel: %d", sl.maxLevel)
			fmt.Println()
			panic(r)
		}
	}()

	// Find the start node to begin overlap check from,
	// i.e. the largest node with an interval less than the query interval.
	n := sl.head
	for i := sl.maxSearchLevel(); i >= 0; i-- {
		for n.levels[i].next != nil && less(n.levels[i].next.intervalKey, interval) {
			n = n.levels[i].next
		}
	}

	// Find overlapping nodes (a < qEnd) && (b > qStart).
	for count := 0; n != nil && n.intervalKey.Start <= interval.End; {
		if n.intervalKey.End >= interval.Start {
			if count >= qParam.Offset {
				result = append(result, &n.intervalKey)
				if qParam.Limit != 0 && len(result) >= qParam.Limit {
					break
				}
			}
			count++
		}
		n = n.levels[0].next
	}
	return result
}

// Get retrieves a key by its interval.
// Returns nil if the interval doesn't exist.
func (sl *SkipList) Get(interval IntervalKey) *IntervalKey {
	n := sl.head
	for i := sl.maxSearchLevel(); i >= 0; i-- {
		for n.levels[i].next != nil && less(n.levels[i].next.intervalKey, interval) {
			n = n.levels[i].next
		}
	}
	n = n.levels[0].next
	if n != nil && n.intervalKey.equalInterval(interval) {
		return &n.intervalKey
	}
	return nil
}

// GetByIndex retrieves a key by its index position in the list.
// The index is 0-based (sl.length < index >= 0 ).
func (sl *SkipList) GetByIndex(index int) (*IntervalKey, error) {
	if index < 0 || index >= sl.length {
		return nil, fmt.Errorf("index out of bounds: %d", index)
	}
	n := sl.head
	pos := index + 1 // Adjust for the head.

	// Traverse from the top level down to the base level.
	for i := sl.maxSearchLevel(); i >= 0; i-- {
		// Move forward at the current level while the remaining position is greater
		// than or equal to the span of the next node.
		for n.levels[i].next != nil && pos >= n.levels[i].span {
			pos -= n.levels[i].span
			n = n.levels[i].next
		}
	}
	if n != nil && n != sl.head {
		return &n.intervalKey, nil
	}
	return nil, fmt.Errorf("node not found at index: %d", index)
}

// Print outputs a visual representation of the list from the given startLevel to the base level.
func (sl *SkipList) Print(w io.Writer, startLevel int) {
	if sl == nil {
		return
	}
	// Clamp the startLevel.
	if startLevel > sl.maxLevel {
		startLevel = sl.maxLevel
	}
	if startLevel < 1 {
		startLevel = 1
	}
	for lvl := startLevel; lvl >= 1; lvl-- {
		fmt.Fprintf(w, "Level %d: ", lvl)
		current := sl.head
		for current != nil {
			fmt.Fprintf(w, "{i: [%d,%d], k: %s, s: %d} -> ",
				current.intervalKey.Start,
				current.intervalKey.End,
				current.intervalKey.Key,
				current.levels[lvl-1].span,
			)
			// Move to the next node at this level
			current = current.levels[lvl-1].next
		}
		fmt.Fprintf(w, "nil\n")
	}
}

// maxSearchLevel returns the effective maximum search limit for level traversal.
// This optimizes performance in large lists by restricting traversal
// to the most relevant lower levels.
func (sl *SkipList) maxSearchLevel() int {
	maxSearchLevel := sl.maxLevel - 1
	if maxSearchLevel >= MaxSearchLevel {
		maxSearchLevel = MaxSearchLevel - 1
	}
	return maxSearchLevel
}
