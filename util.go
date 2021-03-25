package main

import "fmt"

type Iterator interface {
	Reset()
	Next() interface{}
}

type ZipListIterator struct {
	ZL  ZipList
	Idx int
}

func NewZipListIterator(zl ZipList) Iterator {
	it := new(ZipListIterator)
	it.ZL = zl
	return it
}

func (it *ZipListIterator) Next() interface{} {
	if res, err := it.ZL.Get(it.Idx); err != nil {
		fmt.Println("reaches the end")
		return nil
	} else {
		it.Idx += 1
		return res
	}
}

func (it *ZipListIterator) Reset() {
	it.Idx = 0
}

type IntSetIterator struct {
	IS  IntSet
	Idx int
}

func NewIntSetIterator(is IntSet) Iterator {
	it := new(IntSetIterator)
	it.IS = is
	return it
}

func (it *IntSetIterator) Next() (res interface{}) {
	if it.Idx == it.IS.Size() {
		fmt.Println("reaches the end")
		return
	}
	res = it.IS.Get(it.Idx)
	it.Idx++
	return
}

func (it *IntSetIterator) Reset() {
	it.Idx = 0
}

// type QuickListIterator struct {}

// func NewQuickListIterator() Iterator {

// 	it := new(QuickListIterator)

// 	return it
// }
