package islist

import "fmt"

const (
	IntervalEqual    = 0
	IntervalLessThan = -1
	IntervalGreater  = 1
)

// IntervalKey represent a key in the list with an associated interval.
type IntervalKey struct {
	Start, End int64
	Key        string
}

// NewIntervalKey returns a new IntervalKey to insert into a list.
func NewIntervalKey(start, end int64, key string) IntervalKey {
	if start < 0 || end < 0 || (start > end) {
		panic("interval Start and End must be a positive number and Start must be < End")
	}
	return IntervalKey{
		Start: start,
		End:   end,
		Key:   key,
	}
}

// NewQueryInterval returns an IntervalKey used in interval queries.
// This is a convience to avoid having separate types which would require a common interface.
func NewIntervalQuery(start, end int64) IntervalKey {
	return NewIntervalKey(start, end, "")
}

func (i IntervalKey) String() string {
	return fmt.Sprintf("{interval: [%d,%d], key: %s", i.Start, i.End, i.Key)
}

// equalInterval checks if two intervals are considered identical.
func (i IntervalKey) equalInterval(i2 IntervalKey) bool {
	return i.Start == i2.Start && i.End == i2.End
}

// less compares the order of intervals a and b by their Start.
// It compares End if the Start of a and b are equal.
func less(a, b IntervalKey) bool {
	if a.Start < b.Start {
		return true
	}
	if a.Start > b.Start {
		return false
	}
	return a.End < b.End
}
