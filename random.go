package golive

import (
	"crypto/rand"
	"math/big"
	"sync"
)

type Random struct {
	last int64
}

var instantiated *Random
var once sync.Once

func NewLiveID() *Random {
	once.Do(func() {
		instantiated = &Random{}
	})
	return instantiated
}

func (g Random) GenerateSmall() string {
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
