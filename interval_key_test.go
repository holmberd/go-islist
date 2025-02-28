package islist

import "testing"

func TestNewIntervalKeySuccess(t *testing.T) {
	s, e := int64(5), int64(10)
	iv := NewIntervalKey(s, e, "key")
	if iv.Start != 5 || iv.End != e {
		t.Errorf("expected [%d, %d], got %s", s, e, iv)
	}
}

func TestIntervalEqual(t *testing.T) {
	s, e := int64(5), int64(10)
	t.Run("Interval a should equal b", func(t *testing.T) {
		iv1 := NewIntervalKey(s, e, "key1")
		iv2 := NewIntervalKey(s, e, "key2")
		if iv1.equalInterval(iv2) != true {
			t.Errorf("expected %s to equal %s", iv1, iv2)
		}
	})
	t.Run("Interval a should not equal b", func(t *testing.T) {
		iv1 := NewIntervalKey(s, e, "key")
		iv2 := NewIntervalKey(s+1, e, "key")
		if iv1.equalInterval(iv2) == true {
			t.Errorf("expected %s to not equal %s", iv1, iv2)
		}
	})
}

func TestNegativeIntervalPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for negative interval")
		}
	}()
	_ = NewIntervalKey(-5, 10, "key")
}

func TestEndBeforeStartIntervalPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for interval with End before Start")
		}
	}()
	_ = NewIntervalKey(10, 5, "key")
}

func TestLess(t *testing.T) {
	tests := []struct {
		A        IntervalKey
		B        IntervalKey
		Expected bool
	}{
		{IntervalKey{Start: 10, End: 20}, IntervalKey{Start: 25, End: 35}, true},
		{IntervalKey{Start: 24, End: 35}, IntervalKey{Start: 10, End: 20}, false},
		{IntervalKey{Start: 10, End: 20}, IntervalKey{Start: 10, End: 25}, true},
		{IntervalKey{Start: 10, End: 25}, IntervalKey{Start: 10, End: 20}, false},
	}
	for _, test := range tests {
		result := less(test.A, test.B)
		if result != test.Expected {
			t.Errorf("Less(%+v, %+v): expected %t, got %t", test.A, test.B, test.Expected, result)
		}
	}
}
