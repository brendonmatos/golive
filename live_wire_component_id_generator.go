package golive

import (
	"crypto/rand"
	"math/big"
	"sync"
)

type LiveIdGenerator struct {
	last int64
}

var instantiated *LiveIdGenerator
var once sync.Once

func NewLiveId() *LiveIdGenerator {
	once.Do(func() {
		instantiated = &LiveIdGenerator{}
	})
	return instantiated
}

func (g LiveIdGenerator) GenerateRandomString() string {
	a, _ := GenerateRandomString(5)
	return a
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
