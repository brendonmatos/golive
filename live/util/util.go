package util

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"reflect"
	"sync"
)

func JsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1 : len(s)-1]
}

func ReverseSlice(s interface{}) {
	size := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, size-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func CreateUniqueName(name string) string {
	return name + "_" + NewLiveID().GenerateSmall()
}

func RandomSmall() string {
	return NewLiveID().GenerateSmall()
}

type Random struct {
	last int64
}

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

var once sync.Once
var instantiated *Random

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
