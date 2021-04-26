package datastructure

import (
	crand "crypto/rand"
	"encoding/base64"
	"testing"
)

func TestDict_Api(t *testing.T) {
	dict := NewDict()
	strs := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		strs[i] = randstring(10)
	}

	for _, v := range strs {
		dict.Put(v, "a")
	}

	for _, v := range strs {
		if dict.Get(v) == "" {
			panic("no value")
		}
	}

	for _, v := range strs {
		dict.Delete(v)
	}

	for _, v := range strs {
		if dict.Get(v) != "" {
			panic("has value")
		}
	}

	if dict.Size() != 0 {
		panic("has value")
	}
}

func randstring(n int) string {
	b := make([]byte, 2*n)
	crand.Read(b)
	s := base64.URLEncoding.EncodeToString(b)
	return s[0:n]
}

var gs string

func BenchmarkDict(b *testing.B) {
	dict := NewDict()
	strs := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		strs[i] = randstring(10)
	}

	b.ResetTimer()

	for _, v := range strs {
		dict.Put(v, "a")
	}

	for _, v := range strs {
		gs = dict.Get(v)
	}
}

func BenchmarkMap(b *testing.B) {
	m := make(map[string]string)
	strs := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		strs[i] = randstring(10)
	}

	b.ResetTimer()

	for _, v := range strs {
		m[v] = "a"
	}

	for _, v := range strs {
		gs = m[v]
	}
}