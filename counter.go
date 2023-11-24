package main

import (
	"fmt"
	"strings"
	"sync/atomic"
)

const (
	msInDay  = 1000 * 60 * 60 * 24
	msInHour = 1000 * 60 * 60
	msInMin  = 1000 * 60
	msInSec  = 1000
)

type Counter struct {
	counter uint64
}

func NewCounter(initialValue uint64) *Counter {
	c := Counter{}
	atomic.StoreUint64(&c.counter, initialValue)
	return &c
}

func (c *Counter) Inc() {
	atomic.AddUint64(&c.counter, 1*tickInt)
}

func (c *Counter) Load() uint64 {
	return atomic.LoadUint64(&c.counter)
}

func (c *Counter) Disp() string {
	ms := atomic.LoadUint64(&c.counter)
	return (dispTime(ms))
}

func (c *Counter) Reset() {
	atomic.StoreUint64(&c.counter, uint64(0))
}

func dispTime(ms uint64) string {
	disp := []string{}

	numDays := ms / msInDay
	if numDays > 0 {
		disp = append(disp, fmt.Sprintf("%dd", numDays))
	}

	ms = ms % msInDay
	numHours := ms / msInHour
	if numHours > 0 {
		disp = append(disp, fmt.Sprintf("%dh", numHours))
	}

	ms = ms % msInHour
	numMinutes := ms / msInMin
	if numMinutes > 0 {
		disp = append(disp, fmt.Sprintf("%dm", numMinutes))
	}

	ms = ms % msInMin
	numSeconds := ms / msInSec
	ms = ms % msInSec
	disp = append(disp, fmt.Sprintf("%02ds", numSeconds))

	return strings.Join(disp, ":")
}
