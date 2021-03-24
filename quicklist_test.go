package main

import (
	"fmt"
	"testing"
)

func TestQuickList_Main(t *testing.T) {
	l := new(List)

	l.Append(1, 2, 3)

	fmt.Println(l)

	l.Incre()

	fmt.Println(l)
}


type List []int

func (l *List) Incre() {
	
	for i := range(*l) {
		(*l)[i]++
	}
}

func (l *List) Append(nums ...int) {
	
	for _, num := range(nums) {
		*l = append(*l, num)
	}
}