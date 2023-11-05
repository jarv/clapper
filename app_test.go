package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllowedHost(t *testing.T) {
	testCases := []struct {
		host string
		want bool
	}{
		{
			host: "http://jarv.org",
			want: true,
		},
		{
			host: "https://jarv.org",
			want: true,
		},
		{
			host: "http://example.com",
			want: false,
		},
		{
			host: "https://example.com",
			want: false,
		},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.host, func(t *testing.T) {
			assert.Equal(t, tt.want, isAllowedHost(tt.host))
		})
	}
}
