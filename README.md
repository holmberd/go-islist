# go-islist

`go-islist` is a high-performance, in-memory **Interval Skip List** implementation written in Go.

It provides efficient insertion, deletion, and querying for **intervals that overlap a given range**, with logarithmic time complexity under most conditions.

> Skip lists are probabilistic data structures that offer similar performance characteristics to balanced trees but are simpler to implement and reason about.

---

## Features

- **Efficient Overlap Queries** — Fast range lookups for intervals that intersect a target interval.
- **Contiguous Interval Management** — Insert, update, and delete intervals with automatic rebalancing.
- **Indexable** — Retrieve intervals by position in the list.
- **Memory Pooling** — Reuses node memory for reduced GC pressure.
- **Probabilistic Balancing** — Geometric distribution for level assignment.

---

## Package Overview

```go
import "github.com/your-org/go-islist"

sl := islist.New(pool, rand.NewPCG(seed))
sl.Insert(IntervalKey{Start: 10, End: 20, Key: "A"})
sl.Insert(IntervalKey{Start: 30, End: 40, Key: "B"})
overlaps := sl.Overlaps(IntervalKey{Start: 15, End: 35}, islist.QueryParam{Limit: 10})
```

## Usage

### Insert
```go
sl.Insert(IntervalKey{Start: 0, End: 10, Key: "example"})
```

### Delete
```go
sl.Delete(IntervalKey{Start: 0, End: 10, Key: "example"})
```

### Overlap Query
```go
result := sl.Overlaps(
  IntervalKey{Start: 5, End: 15},
  QueryParam{Offset: 0, Limit: 100},
)
```

### Interval Lookup
```go
iv := sl.Get(IntervalKey{Start: 5, End: 15, Key: "foo"})
```

### Index Lookup
```go
iv, err := sl.GetByIndex(3)
```

## Complexity
```
| Operation      | Average Time | Worst Case |
| -------------- | ------------ | ---------- |
| Insert         | O(log n)     | O(n)       |
| Delete         | O(log n)     | O(n)       |
| Overlaps Query | O(log n + k) | O(n)       |
| Index Lookup   | O(log n)     | O(n)       |
```

## Design Notes

### Load Elements
When saving skiplist elements to disk or another data structure, store them in descending order (greatest to smallest). When reloading, insert them into the skiplist in the same descending order.

When elements are inserted in descending order, each new element being inserted is always smaller than the previously inserted element. This means the traversal always stops at the head of the skiplist, and no further traversal is needed.

This avoids the O(log N) traversal, and the insertion completes in O(1).





