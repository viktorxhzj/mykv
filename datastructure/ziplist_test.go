package datastructure

import (
	"fmt"
	"testing"
)

func TestZipList_PushAndElementAt(t *testing.T) {
	z := NewZipList()

	sli := []interface{}{"Hello", "分布式小盒子", 0, 13, -1, 14, -1 << 7, 1<<7 - 1, -1<<7 - 10, 1<<7 + 10, -1 << 15, 1<<15 - 1, -1<<15 - 10, 1<<15 + 10, -1 << 31, 1<<31 - 1, -1<<31 - 10, 1<<31 + 10, -1 << 63, 1<<63 - 1}

	for i := 0; i < len(sli); i++ {
		if ii, ok := sli[i].(int); ok {
			if err := z.AddInt(ii); err != nil {
				fmt.Println(err)
			}
		}
		if ss, ok := sli[i].(string); ok {
			if err := z.AddString(ss); err != nil {
				fmt.Println(err)
			}
		}
	}

	got := int(z.ZLLen())
	fmt.Printf("ZipList.ZLLen() = %v, valid=%v\n", got, got == len(sli))

	sli2 := []interface{}{"Hello", "分布式盒子", 0, 13, -1, 14, -1 << 7, 1<<7 - 1, -1<<7 - 10, 1<<7 + 10, -1 << 15, 1<<15 - 1, -1<<15 - 10, 1<<15 + 10, -1 << 31, 1<<31 - 1, -1<<31 - 10, 1<<31 + 10, -1 << 63, 1<<63 - 1}

	for i := 0; i < len(sli2); i++ {
		if find, err := z.Get(i); err != nil {
			fmt.Println(err)
		} else {
			if ii, ok := find.(int); ok {
				fmt.Printf("ZipList.Find() = %v\n", ii)
			} else {
				fmt.Printf("ZipList.Find() = %v\n", find.(string))
			}

		}
	}
}
