package islist

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"
)

const (
	iRange       = 1_000_000
	iSpan        = 10
	numIntervals = iRange / iSpan
)

const (
	tenMinutes = 10
	oneHour    = 60
	oneDay     = 24 * 60      // 1440 minutes
	oneWeek    = 7 * oneDay   // 10080 minutes
	oneMonth   = 30 * oneDay  // 43200 minutes
	oneYear    = 365 * oneDay // 525600 minutes
)

var randIntervals = make([]IntervalKey, 0, numIntervals)

// init initializes the benchmark tests.
func init() {
	var start int64
	for start = 0; start < iRange; start += iSpan + 1 {
		randIntervals = append(randIntervals, NewIntervalKey(start, start+iSpan, "key"))
	}
	rand.Shuffle(len(randIntervals), func(i, j int) {
		randIntervals[i], randIntervals[j] = randIntervals[j], randIntervals[i]
	})
}

func newRandomInterval(bound int64, length int64) IntervalKey {
	start := rand.Int63n(bound)
	end := start + rand.Int63n(length) + 1
	return NewIntervalKey(start, end, "key")
}

func newPopulatedTestList() (*SkipList, []IntervalKey) {
	list := newTestList()
	intervals := make([]IntervalKey, numIntervals)
	copy(intervals, randIntervals)
	for _, i := range intervals {
		list.Insert(i)
	}
	return list, intervals
}

// chooseOperation selects an operation based on weighted percentages.
func chooseOperation(ops map[string]float64) string {
	r := rand.Float64()
	cumulative := 0.0
	for op, weight := range ops {
		cumulative += weight
		if r < cumulative {
			return op
		}
	}
	return "" // Shouldn't reach here if percentages sum to 1.0.
}

// 1 year = 525600 minutes. 10-minute interval slots = 52560 intervals.
func generateYearIntervals10Min() []IntervalKey {
	is := make([]IntervalKey, 0, oneYear/tenMinutes)
	var start int64
	for start = 0; start < oneYear; start += tenMinutes {
		end := start + tenMinutes
		is = append(is, NewIntervalKey(start, end, "key"))
	}
	return is
}

func BenchmarkISListAscendingInsert(b *testing.B) {
	list := newTestList()
	b.ReportAllocs()
	b.ResetTimer()
	var i int64
	for i = 0; i < int64(b.N); i += 11 {
		list.Insert(NewIntervalKey(i, i+10, "key"))
	}
}

func BenchmarkISListDescendingInsert(b *testing.B) {
	list := newTestList()
	b.ReportAllocs()
	b.ResetTimer()
	var i int64
	for i = int64(b.N * 11); i >= 0; i -= 11 {
		list.Insert(NewIntervalKey(i, i+10, "key"))
	}
}

func BenchmarkISListRandomInsert(b *testing.B) {
	list := newTestList()
	intervals := make([]IntervalKey, numIntervals)
	copy(intervals, randIntervals)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Insert(intervals[i%len(intervals)])
	}
}

func BenchmarkISListDelete(b *testing.B) {
	list, intervals := newPopulatedTestList()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Delete(intervals[i%len(intervals)])
	}
}

func BenchmarkISListOverlaps(b *testing.B) {
	list, _ := newPopulatedTestList()
	querySpans := []int64{10, 100, 1000}
	for _, span := range querySpans {
		b.Run(fmt.Sprintf("Query Span=%d", span), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = list.Overlaps(newRandomInterval(numIntervals, span), QueryParam{})
			}
		})
	}
}

// Benchmark mixed operations throughput.
func BenchmarkISListMixedOperations(b *testing.B) {
	ops := map[string]float64{
		"insert":   0.2,
		"delete":   0.1,
		"overlaps": 0.7,
	}
	ranges := []int64{200_000, 2_000_000}
	for _, r := range ranges {
		b.Run(fmt.Sprintf("Range=%d", r), func(b *testing.B) {
			list := newTestList()
			addedIntervals := make([]IntervalKey, 0, b.N)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				operation := chooseOperation(ops)
				interval := newRandomInterval(r, iSpan)

				switch operation {
				case "insert":
					list.Insert(interval)
					addedIntervals = append(addedIntervals, interval)
				case "delete":
					if len(addedIntervals) > 0 {
						index := rand.Intn(len(addedIntervals))
						list.Delete(addedIntervals[index])
						addedIntervals = slices.Delete(addedIntervals, index, index+1)
					}
				case "overlaps":
					_ = list.Overlaps(interval, QueryParam{})
				}
			}
		})
	}
}

func BenchmarkISListTimeSpanOverlapQueries(b *testing.B) {
	list := newTestList()
	for _, iv := range generateYearIntervals10Min() {
		list.Insert(iv)
	}
	generateQuery := func(maxLen int64) IntervalKey {
		start := rand.Int63n(oneYear)
		length := rand.Int63n(maxLen)
		end := start + length + 1
		return NewIntervalKey(start, end, "key")
	}
	timeSpans := []struct {
		name string
		span int64
	}{
		{"10min", tenMinutes},
		{"1h", oneHour},
		{"1d", oneDay},
		{"1w", oneWeek},
		{"1m", oneMonth},
		{"3m", 3 * oneMonth},
	}
	for _, ts := range timeSpans {
		b.Run(fmt.Sprintf("Query Span=%s", ts.name), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = list.Overlaps(generateQuery(ts.span), QueryParam{})
			}
		})
	}
}
