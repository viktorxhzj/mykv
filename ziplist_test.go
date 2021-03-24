package main

// import (
// 	"fmt"
// 	"testing"
// )

// func TestZipList_PushAndElementAt(t *testing.T) {
// 	z := NewZipList()

// 	sli := []interface{}{[]byte{0x12, 0xff, 0xed}, "Hello", "分布式小盒子", 0, 13, -1, 14, -1 << 7, 1<<7 - 1, -1<<7 - 10, 1<<7 + 10, -1 << 15, 1<<15 - 1, -1<<15 - 10, 1<<15 + 10, -1 << 31, 1<<31 - 1, -1<<31 - 10, 1<<31 + 10, -1 << 63, 1<<63 - 1}

// 	for i := 0; i < len(sli); i++ {
// 		z.Add(sli[i])
// 	}

// 	got := int(z.ZLLen())
// 	fmt.Printf("ZipList.ZLLen() = %v, valid=%v\n", got, got == len(sli))

// 	sli2 := []interface{}{[]byte{0x12, 0xff, 0xed}, "Hello", "分布式盒子", 0, 13, -1, 14, -1 << 7, 1<<7 - 1, -1<<7 - 10, 1<<7 + 10, -1 << 15, 1<<15 - 1, -1<<15 - 10, 1<<15 + 10, -1 << 31, 1<<31 - 1, -1<<31 - 10, 1<<31 + 10, -1 << 63, 1<<63 - 1}

// 	for i := 0; i < len(sli2); i++ {
// 		find := z.Find(sli2[i])
// 		fmt.Printf("ZipList.Find() = %v\n", find)
// 	}
// }

// func TestZipList_Iterator(t *testing.T) {
// 	z := NewZipList()

// 	sli := []interface{}{[]byte{0x12, 0xff, 0xed}, "Hello", "分布式小盒子", 0, 13, -1, 14, -1 << 7, 1<<7 - 1, -1<<7 - 10, 1<<7 + 10, -1 << 15, 1<<15 - 1, -1<<15 - 10, 1<<15 + 10, -1 << 31, 1<<31 - 1, -1<<31 - 10, 1<<31 + 10, -1 << 63, 1<<63 - 1}

// 	for i := 0; i < len(sli); i++ {
// 		z.Add(sli[i])
// 	}

// 	it := NewZipListIterator(z)

// 	for {
// 		nxt := it.Next()
// 		if nxt == nil {
// 			break
// 		}
// 		fmt.Println(nxt)
// 	}
// }
