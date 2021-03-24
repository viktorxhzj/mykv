package main

import (
	"fmt"
	"math"
	"testing"
)

func TestIntSetFind(t *testing.T) {
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

func TestIntSetUpgrade(t *testing.T) {
	is := NewIntSet()

	sli := []int{math.MinInt16, math.MaxInt16, math.MinInt32, math.MaxInt32, math.MinInt64, math.MaxInt64}

	for _, v := range sli {
		is.Add(v)
		fmt.Printf("%x\n", is)
	}

	for i := range sli {
		fmt.Println(is.Get(i))
	}

	for _, v := range sli {
		_, exists := is.Find(v)
		fmt.Println(exists)
	}
}
