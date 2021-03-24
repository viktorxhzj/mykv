package main

import (
	"fmt"
	"testing"
)

func TestZipList_PushAndElementAt(t *testing.T) {
	z := NewZipList()

	sli := []interface{}{"Hello", "肖何子建", 0, 13, -1, 14, -1 << 7, 1<<7 - 1, -1<<7 - 10, 1<<7 + 10, -1 << 15, 1<<15 - 1, -1<<15 - 10, 1<<15 + 10, -1 << 31, 1<<31 - 1, -1<<31 - 10, 1<<31 + 10, -1 << 63, 1<<63 - 1}

	for i := 0; i < len(sli); i++ {
		z.Push(sli[i])
	}

	got := int(z.ZLLen())
	fmt.Printf("ZipList.ZLLen() = %v, valid=%v\n", got, got == len(sli))

	for i := 0; i < len(sli); i++ {
		got, _ := z.ElementAt(i)
		fmt.Printf("ZipList.ElementAt() = %v,valid=%v\n", got, got == sli[i])
	}
}
