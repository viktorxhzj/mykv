package datastructure

import (
	"fmt"
	"testing"
)

func TestQuickList_PushHead(t *testing.T) {
	list := NewQuickList()
	sli := []interface{}{"Hello", "分布式小盒子", 0, 13, -1, 14, -1 << 7, 1<<7 - 1, -1<<7 - 10, 1<<7 + 10, -1 << 15, 1<<15 - 1, -1<<15 - 10, 1<<15 + 10, -1 << 31, 1<<31 - 1, -1<<31 - 10, 1<<31 + 10, -1 << 63, 1<<63 - 1}

	for i := range sli {
		list.PushHead(sli[i])
	}

	for i := range sli {
		e, _ := list.Get(i)
		fmt.Println(e)
	}
}

func TestQuickList_PushTail(t *testing.T) {
	list := NewQuickList()
	sli := []interface{}{"Hello", "分布式小盒子", 0, 13, -1, 14, -1 << 7, 1<<7 - 1, -1<<7 - 10, 1<<7 + 10, -1 << 15, 1<<15 - 1, -1<<15 - 10, 1<<15 + 10, -1 << 31, 1<<31 - 1, -1<<31 - 10, 1<<31 + 10, -1 << 63, 1<<63 - 1}

	for i := range sli {
		list.PushTail(sli[i])
	}

	for i := range sli {
		e, _ := list.Get(i)
		fmt.Println(e)
	}
}
