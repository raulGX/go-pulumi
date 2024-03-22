package main

import (
	"testing"
)

func TestRingBuffer(t *testing.T) {
	r := newRingBuffer(3)
	r.Update(9)
	r.Update(5)
	r.Update(1)
	if r.Average() != 5 {
		t.Errorf("got %f, want 5", r.Average())
	}

	r.Update(3)
	if r.Average() != 3 {
		t.Errorf("got %f, want 3", r.Average())
	}

	r.Update(4)
	r.Update(5)
	if r.Average() != 4 {
		t.Errorf("got %f, want 4", r.Average())
	}
}
