package gofss

import (
	"testing"
)

var result string

func BenchmarkHash8(b *testing.B) {
	var r string
	for n := 0; n < b.N; n++ {
		r = NewHash(8)
	}
	result = r
}
