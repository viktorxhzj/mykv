package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestNewSkipList_Main(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	list := NewSkipList()

	type A struct {
		Key string
		Val float64
	}

	k := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	v := []float64{0.68, 0.47, 0.57, 0.69, 0.59, 0.42, 0.29, 0.24, 0.43, 0.92}

	for i := range k {
		v[i] = rand.Float64()
	}

	for i := range k {
		list.Add(k[i], v[i])
	}


	for i := range k {
		fmt.Print(list.Contains(k[i], v[i]), " ")
	}
	
	for i := range k {
		fmt.Print(list.GetRank(k[i], v[i]), " ")
	}


}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}
