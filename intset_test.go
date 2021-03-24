package main

import (
	"fmt"
	"math"
	"testing"
)

func TestIntSet_Find(t *testing.T) {
	is := NewIntSet()

	sli := []int{1, 2, 3}

	for _, v := range sli {
		is.Add(v)
	}

	for _, v := range sli {
		_, exists := is.Find(v)
		fmt.Println(exists)

		_, exists = is.Find(v + 10)
		fmt.Println(exists)
	}
}

func TestIntSet_Upgrade(t *testing.T) {
	var is IntSet = NewIntSet()

	sli := []int{math.MinInt16, math.MaxInt16, math.MinInt32, math.MaxInt32, math.MinInt64, math.MaxInt64}

	for _, v := range sli {
		is.Add(v)
		fmt.Printf("%x\n", is)
	}

	for i := range sli {
		fmt.Println(is.Get(i))
	}

	it := NewIntSetIterator(is)

	for {
		nxt := it.Next()
		if nxt == nil {
			break
		}
		fmt.Println(nxt)
	}

	for _, v := range sli {
		_, exists := is.Find(v)
		fmt.Println(exists)
	}
}

func TestIntSet_Iterator(t *testing.T) {
	var is IntSet = NewIntSet()

	sli := []int{math.MinInt16, math.MaxInt16, math.MinInt32, math.MaxInt32, math.MinInt64, math.MaxInt64}

	for _, v := range sli {
		is.Add(v)
		fmt.Printf("%x\n", is)
	}

	for i := range sli {
		fmt.Println(is.Get(i))
	}

	it := NewIntSetIterator(is)

	for {
		nxt := it.Next()
		if nxt == nil {
			break
		}
		fmt.Println(nxt)
	}
}
