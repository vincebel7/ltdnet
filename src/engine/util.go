package engine

import (
	"math/rand"
	"time"
)

func IDgen(n int) string {
	var idchars = []rune("abcdef1234567890")
	id := make([]rune, n)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range id {
		id[i] = idchars[r.Intn(len(idchars))]
	}

	return string(id)
}
