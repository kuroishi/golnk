package idgen

import "crypto/rand"

const idChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type CryptoGen struct{}

func NewCryptoGen() *CryptoGen {
	return &CryptoGen{}
}

func (g *CryptoGen) Generate() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i, v := range b {
		b[i] = idChars[int(v)%len(idChars)]
	}
	return string(b)
}