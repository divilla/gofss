package gofss

import (
	"crypto/rand"
	"math/big"
)

var URL64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

func NewHash(size int) string {
	var word string

	for i := 0; i < size; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(64*64*64*64*64*64*64*64*64*64))
		if err != nil {
			panic(err)
		}
		n := nBig.Int64()

		for j := 0; j < 10; j++ {
			m := n % 64
			word += URL64[m : m+1]
			n = n / 64
		}
	}

	return word
}
