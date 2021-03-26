package main

import "fmt"

type Iterator interface {
	Reset()
	Next() interface{}
}

func AssertValidType(e interface{}) (ss []byte, ii, t int) {
	if s, ok := e.(string); ok {
		ss = []byte(s)
	} else if i, ok := e.(int); ok {
		ii = i
		t = 1
	} else {
		t = -1
	}
	return
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
	res, _ = it.IS.Get(it.Idx)
	it.Idx++
	return
}

func (it *IntSetIterator) Reset() {
	it.Idx = 0
}

type QuickListIterator struct {
	QL  QuickList
	Idx int
}

func NewQuickListIterator() Iterator {

	it := new(QuickListIterator)

	return it
}

func (it *QuickListIterator) Next() (res interface{}) {
	if it.Idx == it.QL.Size() {
		fmt.Println("reaches the end")
		return
	}

	res, _ = it.QL.Get(it.Idx)
	it.Idx++
	return
}

func (it *QuickListIterator) Reset() {
	it.Idx = 0
}
