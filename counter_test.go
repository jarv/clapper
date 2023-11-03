package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	testCases := []struct {
		name         string
		initialValue uint64
		numAdds      int
		want         uint64
	}{
		{
			name:         "add once",
			initialValue: 0,
			numAdds:      1,
			want:         1,
		},
		{
			name:         "add multiple",
			initialValue: 0,
			numAdds:      9000,
			want:         9000,
		},
		{
			name:         "add with initial value",
			initialValue: 100,
			numAdds:      9000,
			want:         9100,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := NewCounter(tt.initialValue)
			for i := 0; i < tt.numAdds; i++ {
				c.Inc()
			}
			assert.Equal(t, c.Load(), tt.want)
		})
	}
}

func TestLoad(t *testing.T) {
	for _, i := range []uint64{0, 100, 300, 9000} {
		c := NewCounter(i)
		assert.Equal(t, c.Load(), i)
	}
}

func TestDisp(t *testing.T) {
	testCases := []struct {
		initialValue uint64
		want         string
	}{
		{
			initialValue: (4783181 * 1000),
			want:         "55d:8h:39m:41.0s",
		},
		{
			initialValue: (4210 * 1000),
			want:         "1h:10m:10.0s",
		},
		{
			initialValue: (190 * 1000),
			want:         "3m:10.0s",
		},
		{
			initialValue: (12 * 1000),
			want:         "12.0s",
		},
		{
			initialValue: 500,
			want:         "00.5s",
		},
		{
			initialValue: 0,
			want:         "00.0s",
		},
		{
			initialValue: 100,
			want:         "00.1s",
		},
		{
			initialValue: 985,
			want:         "00.9s",
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			c := NewCounter(tt.initialValue)
			assert.Equal(t, tt.want, c.Disp())
		})
	}
}

func TestReset(t *testing.T) {
	testCases := []struct {
		initialValue uint64
		numInc       int
	}{
		{
			initialValue: 0,
			numInc:       0,
		},
		{
			initialValue: 999,
			numInc:       0,
		},
		{
			initialValue: 50,
			numInc:       9000,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(fmt.Sprintf("InitialValue: %v NumInc: %v", tt.initialValue, tt.numInc), func(t *testing.T) {
			t.Parallel()
			c := NewCounter(tt.initialValue)
			for i := 0; i < tt.numInc; i++ {
				c.Inc()
			}
			c.Reset()
			assert.Equal(t, uint64(0), c.Load())
			assert.Equal(t, "00.0s", c.Disp())
		})
	}
}
