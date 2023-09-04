package toolkit

import (
	"crypto/rand"
	"log"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

type Tools struct {
}

func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)

	rLen := len(r)
	uIntRlen := uint64(len(r))

	for i := range s {
		p, err := rand.Prime(rand.Reader, rLen)

		if err != nil {
			log.Printf("An error occurred: %v", err)
		}

		x := p.Uint64()
		s[i] = r[x%uIntRlen]
	}

	return string(s)
}
