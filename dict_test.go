package main

import (
	"fmt"
	"testing"
)



func TestDict_Put(t *testing.T) {
	d := NewDict()

	keys := []interface{}{"Hello", "分布式小盒子", 0, 13, -1, 14, -1 << 7, 1<<7 - 1, -1<<7 - 10, 1<<7 + 10, -1 << 15, 1<<15 - 1, -1<<15 - 10, 1<<15 + 10, -1 << 31, 1<<31 - 1, -1<<31 - 10, 1<<31 + 10, -1 << 63, 1<<63 - 1}

	for i := range keys {
		d.Put(keys[i], keys[i])
	}

	fmt.Println(len(keys), d.Size())

	for _, k := range keys {
		fmt.Println(d.Get(k))
	}
}