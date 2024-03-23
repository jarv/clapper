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

type Clapper struct {
	clapper uint64
}

func NewClapper(initialValue uint64) *Clapper {
	c := Clapper{}
	atomic.StoreUint64(&c.clapper, initialValue)
	return &c
}

func (c *Clapper) Inc() {
	atomic.AddUint64(&c.clapper, 1*tickInt)
}

func (c *Clapper) Load() uint64 {
	return atomic.LoadUint64(&c.clapper)
}

func (c *Clapper) Disp() string {
	ms := atomic.LoadUint64(&c.clapper)
	return (dispTime(ms))
}

func (c *Clapper) Reset() {
	atomic.StoreUint64(&c.clapper, uint64(0))
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
	disp = append(disp, fmt.Sprintf("%ds", numSeconds))

	return strings.Join(disp, ":")
}
