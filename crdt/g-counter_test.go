package crdt

import "testing"

func Test(t *testing.T) {
	a := NewGCounter("node1")
	b := NewGCounter("node2")

	a.Increment()
	b.Increment()
	a.Increment()
	b.Increment()
	a.Increment()

	a.Merge(*b)

	if a.Value() != 5 {
		t.Errorf("a.Value() = %d, want 5", a.Value())
	}

	b.Merge(*a)

	if a.Value() != 5 {
		t.Errorf("a.Value() = %d, want 5", a.Value())
	}

}
