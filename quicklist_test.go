package main

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestQuickList_Main(t *testing.T) {
	z := zipListEntry{}
	fmt.Println(unsafe.Sizeof(z))
}

type List []int

func (l *List) Incre() {

	for i := range *l {
		(*l)[i]++
	}
}

func (l *List) Append(nums ...int) {

	for _, num := range nums {
		*l = append(*l, num)
	}
}
