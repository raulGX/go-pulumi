package main

import (
	"errors"
	"sync"
)

type Tracker interface {
	Value() (v float32, err error)
	Average() float32
	Update(v float32)
}

var (
	ErrIsEmpty = errors.New("ringbuffer is empty")
)

type ringBuffer struct {
	buffer      []float32
	maxSize     uint
	average     float32
	isFull      bool
	readOffset  uint
	writeOffset uint
	m           sync.Mutex
}

func (r *ringBuffer) Value() (v float32, err error) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.readOffset == r.writeOffset {
		return 0, ErrIsEmpty
	}
	return r.buffer[r.readOffset], nil
}
func (r *ringBuffer) Average() float32 {
	return r.average
}
func (r *ringBuffer) setAverage() {
	start := uint(0)
	end := r.maxSize - 1
	if !r.isFull {
		end = r.readOffset
	}
	var sum float32 = 0
	for i := start; i <= end; i++ {
		sum += r.buffer[i]
	}
	r.average = sum / float32(end-start+1)
}
func (r *ringBuffer) Update(v float32) {
	r.m.Lock()
	defer r.m.Unlock()
	r.buffer[r.writeOffset] = v
	r.writeOffset = (r.writeOffset + 1) % r.maxSize
	if r.writeOffset == 0 {
		r.isFull = true
		r.readOffset = r.maxSize - 1
	} else {
		r.readOffset = r.writeOffset - 1
	}
	r.setAverage()
}

func newRingBuffer(size uint) Tracker {
	return &ringBuffer{
		buffer:     make([]float32, size),
		maxSize:    size,
		average:    0,
		isFull:     false,
		readOffset: 0, writeOffset: 0,
	}
}
