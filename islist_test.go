package islist

import (
	"math/rand/v2"
	"testing"
)

// TODO: Add no-overlap test.

// Helper function to create a test tree.
func newTestList() *SkipList {
	return New(NewNodePool(), rand.NewPCG(2, 3))
}

type expectedNode struct {
	intervalKey IntervalKey
	span        int
}

func assertNodeEqual(t *testing.T, level int, n *Node, expected expectedNode) {
	if n == nil {
		t.Fatalf("expected node but got nil")
	}
	if !n.intervalKey.equalInterval(expected.intervalKey) {
		t.Errorf("node interval mismatch. got %s, expected %s", n.intervalKey, expected.intervalKey)
	}
	if n.intervalKey.Key != expected.intervalKey.Key {
		t.Errorf("node key mismatch. got %s, expected %s", n.intervalKey.Key, expected.intervalKey.Key)
	}
	if n.levels[level].span != expected.span {
		t.Errorf("node span mismatch. got %d, expected %d", n.levels[level].span, expected.span)
	}
}

func assertNodesEqual(t *testing.T, sl *SkipList, expectedNodes []expectedNode) {
	t.Helper()
	var n *Node
	for i, j := sl.maxLevel-1, 0; i >= 0; i-- {
		n = sl.head
		for n != nil {
			assertNodeEqual(t, i, n, expectedNodes[j])
			j++
			n = n.levels[i].next
		}
	}
}

type expectedList struct {
	level  int
	length int
}

func assertListEqual(t *testing.T, sl *SkipList, expected expectedList) {
	t.Helper()
	if sl.maxLevel != expected.level {
		t.Errorf("list level mismatch. got %d, expected %d", sl.maxLevel, expected.level)
	}
	if sl.length != expected.length {
		t.Errorf("list length mismatch. got %d, expected %d", sl.length, expected.length)
	}
}

func TestInsert(t *testing.T) {
	t.Run("Insert intervals", func(t *testing.T) {
		list := newTestList()
		list.Insert(NewIntervalKey(5, 9, "test-1"))
		list.Insert(NewIntervalKey(10, 20, "test-2"))
		list.Insert(NewIntervalKey(30, 40, "test-3"))
		list.Insert(NewIntervalKey(50, 60, "test-4"))
		assertListEqual(t, list, expectedList{level: 3, length: 4})
		assertNodesEqual(t, list, []expectedNode{
			// Level 3: [0,0] -> [10,20] -> nil
			{intervalKey: NewIntervalKey(0, 0, ""), span: 2},
			{intervalKey: NewIntervalKey(10, 20, "test-2"), span: 2},
			// Level 2: [0,0] -> [10,20] -> [50,60] -> nil
			{intervalKey: NewIntervalKey(0, 0, ""), span: 2},
			{intervalKey: NewIntervalKey(10, 20, "test-2"), span: 2},
			{intervalKey: NewIntervalKey(50, 60, "test-4"), span: 0},
			// Level 1: [0,0] -> [5,9] -> [10,20] -> [30,40] -> [50,60] -> nil
			{intervalKey: NewIntervalKey(0, 0, ""), span: 1},
			{intervalKey: NewIntervalKey(5, 9, "test-1"), span: 1},
			{intervalKey: NewIntervalKey(10, 20, "test-2"), span: 1},
			{intervalKey: NewIntervalKey(30, 40, "test-3"), span: 1},
			{intervalKey: NewIntervalKey(50, 60, "test-4"), span: 0},
		})
	})

	t.Run("Update interval key", func(t *testing.T) {
		list := newTestList()
		key1 := "original"
		key2 := "updated"
		k := list.Insert(NewIntervalKey(10, 20, key1))
		if k != nil {
			t.Errorf("expected empty key returned. got %s", k)
		}
		k = list.Insert(NewIntervalKey(10, 20, key2))
		if k.Key != key1 {
			t.Errorf("expected old key %s returned. got %s", key1, k)
		}
	})
}

func TestOverlapQuery(t *testing.T) {
	list := newTestList()
	list.Insert(NewIntervalKey(5, 9, "test-1"))
	list.Insert(NewIntervalKey(10, 20, "test-2"))
	list.Insert(NewIntervalKey(30, 40, "test-3"))
	list.Insert(NewIntervalKey(50, 60, "test-4"))
	list.Insert(NewIntervalKey(70, 80, "test-5"))
	list.Insert(NewIntervalKey(90, 100, "test-6"))

	t.Run("Overlap query end before list minStart", func(t *testing.T) {
		r := list.Overlaps(NewIntervalQuery(1, 2), QueryParam{})
		if len(r) != 0 {
			t.Errorf("expected no overlapping intervals. got %d", len(r))
		}
	})

	t.Run("Overlap query end after list maxEnd", func(t *testing.T) {
		r := list.Overlaps(NewIntervalQuery(101, 102), QueryParam{})
		if len(r) != 0 {
			t.Errorf("expected no overlapping intervals. got %d", len(r))
		}
	})

	t.Run("Overlap query start before list minStart", func(t *testing.T) {
		r := list.Overlaps(NewIntervalQuery(1, 14), QueryParam{})
		if len(r) != 2 {
			t.Errorf("expected 2 overlapping intervals. got %d", len(r))
		}
	})

	t.Run("Overlap query end after list maxEnd", func(t *testing.T) {
		r := list.Overlaps(NewIntervalQuery(75, 105), QueryParam{})
		if len(r) != 2 {
			t.Errorf("expected 2 overlapping intervals. got %d", len(r))
		}
	})

	t.Run("Overlap query start before list minStart and end after maxEnd", func(t *testing.T) {
		r := list.Overlaps(NewIntervalQuery(1, 105), QueryParam{})
		if len(r) != list.length {
			t.Errorf("expected %d overlapping intervals. got %d", list.length, len(r))
		}
	})

	t.Run("Overlap query between two intervals", func(t *testing.T) {
		r := list.Overlaps(NewIntervalQuery(21, 25), QueryParam{})
		if len(r) != 0 {
			t.Errorf("expected 0 overlapping intervals. got %d", len(r))
		}
	})

	t.Run("Overlap query inclusive start and inclusive end", func(t *testing.T) {
		r := list.Overlaps(NewIntervalQuery(40, 50), QueryParam{})
		if len(r) != 2 {
			t.Errorf("expected 2 overlapping intervals. got %d", len(r))
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("Delete last interval in list", func(t *testing.T) {
		delete := NewIntervalKey(10, 20, "test-2")
		list := newTestList()
		list.Insert(delete)
		assertListEqual(t, list, expectedList{level: 1, length: 1})
		k := list.Delete(delete)
		if k == nil {
			t.Errorf("expected key after delete, got %s", k)
		}
		assertListEqual(t, list, expectedList{level: 1, length: 0})
	})
	t.Run("Delete last top level interval should adjust list maxLevel", func(t *testing.T) {
		delete := NewIntervalKey(10, 20, "test-2")
		list := newTestList()
		list.Insert(NewIntervalKey(5, 9, "test-1"))
		list.Insert(delete)
		list.Insert(NewIntervalKey(30, 40, "test-3"))
		list.Insert(NewIntervalKey(50, 60, "test-4"))
		assertListEqual(t, list, expectedList{level: 3, length: 4})
		k := list.Delete(delete)
		if k == nil {
			t.Errorf("expected key after delete, got %s", k)
		}
		assertListEqual(t, list, expectedList{level: 2, length: 3})
		assertNodesEqual(t, list, []expectedNode{
			// Level 2: [0,0] -> [50,60] -> nil
			{intervalKey: NewIntervalKey(0, 0, ""), span: 3},
			{intervalKey: NewIntervalKey(50, 60, "test-4"), span: 0},
			// Level 1: [0,0] ->  [5,9] -> [30,40] -> [50,60] -> nil
			{intervalKey: NewIntervalKey(0, 0, ""), span: 1},
			{intervalKey: NewIntervalKey(5, 9, "test-1"), span: 1},
			{intervalKey: NewIntervalKey(30, 40, "test-3"), span: 1},
			{intervalKey: NewIntervalKey(50, 60, "test-4"), span: 0},
		})
	})

	t.Run("Delete non-existing interval in empty list", func(t *testing.T) {
		list := newTestList()
		assertListEqual(t, list, expectedList{level: 1, length: 0})
		k := list.Delete(NewIntervalQuery(1, 2))
		assertListEqual(t, list, expectedList{level: 1, length: 0})
		if k != nil {
			t.Errorf("expected nil returned after delete")
		}
	})

	t.Run("Delete non-existing interval in populated list", func(t *testing.T) {
		delete := NewIntervalKey(10, 20, "test-5")
		list := newTestList()
		list.Insert(NewIntervalKey(5, 9, "test-1"))
		list.Insert(NewIntervalKey(30, 40, "test-2"))
		list.Insert(NewIntervalKey(50, 60, "test-4"))
		assertListEqual(t, list, expectedList{level: 3, length: 3})
		k := list.Delete(delete)
		assertListEqual(t, list, expectedList{level: 3, length: 3})
		if k != nil {
			t.Errorf("expected nil returned after delete")
		}
	})
}
