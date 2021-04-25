package innerstructure

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"
)

func TestDict_Put(t *testing.T) {
	d := NewDict()

	keys := []string{"Hello", "分布式小盒子"}

	for i := range keys {
		d.Put(keys[i], keys[i])
	}

	fmt.Println(len(keys), d.Size())

	for _, k := range keys {
		fmt.Println(d.Get(k))
	}
}

func randstring(n int) string {
	b := make([]byte, 2*n)
	crand.Read(b)
	s := base64.URLEncoding.EncodeToString(b)
	return s[0:n]
}

func BenchmarkMap(b *testing.B) {
	dict := NewDict()
	strs := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		strs[i] = randstring(10)
	}

	b.ResetTimer()

	for _, v := range strs {
		dict.Put(v, 1)
	}

	for _, v := range strs {
		if dict.Get(v).(int) != 1 {
			panic("not have")
		}
	}
}

func BenchmarkDict(b *testing.B) {
	m := make(map[string]int)
	strs := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		strs[i] = randstring(10)
	}

	b.ResetTimer()

	for _, v := range strs {
		m[v] = 1
	}

	for _, v := range strs {
		if m[v] != 1 {
			panic("not have")
		}
	}
}