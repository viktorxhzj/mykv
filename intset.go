package main

import (
	"fmt"
	"math"
)

const (
	INTSET_ENC_INT16 uint8 = 2
	INTSET_ENC_INT32 uint8 = 4
	INTSET_ENC_INT64 uint8 = 8
)

// IntSet has a maximum length of UINT32_MAX
// Padding:
// |Encoding 	|Length |Contents						|
// |XOOO 		|XXXX 	|XXXX XXXX XXXX XXXX XXXX XXXX	|
type IntSet struct {
	Encoding uint8
	Length   uint32
	Contents []uint8
}

func NewIntSet() *IntSet {
	is := new(IntSet)
	is.Encoding = INTSET_ENC_INT16
	return is
}

// Find searches for the position of "value".
// Returns true when the value was found and
// sets "pos" to the position of the value within the intset.
// Returns false when the value is not present in the intset
// and sets "pos" to the position where "value" can be inserted.
func (is *IntSet) Find(n int) (int, bool) {
	length := int(is.Length)
	if length == 0 {
		return 0, false
	}

	mi, ma := is.Get(0), is.Get(length-1)
	if n < mi {
		return 0, false
	}
	if n > ma {
		return length, false
	}

	l, r := 0, length-1

	for l < r {
		m := l + (r-l)>>1
		if is.Get(m) == n {
			return m, true
		} else if is.Get(m) < n {
			l = m + 1
		} else {
			r = m - 1
		}
	}

	if is.Get(l) == n {
		return l, true
	} else if is.Get(l) > n {
		return l, false
	}
	return l + 1, false
}

func (is *IntSet) Add(n int) (success bool) {

	if is.Length == math.MaxInt32 {
		fmt.Println("intset reaches the maximum length")
		return
	}

	// 1. get encoding
	enc := intsetValueEncoding(n)

	// if encoding needs upgrade
	if enc > is.Encoding {
		is.upgradeAndAdd(n)
		return true
	} else {
		// abort if already in the set
		if idx, exists := is.Find(n); exists {
			fmt.Println("input already exists in the intset")
			return
		} else {
			is.resize(int(is.Length) + 1)
			is.moveTail(idx)
			is.Set(n, idx)
			is.Length++
			return true
		}
	}
}

func (is *IntSet) Set(n, idx int) {
	if idx < 0 || idx > int(is.Length) {
		fmt.Println("invalid input idx")
		return
	}

	offset := idx * int(is.Encoding)
	var b []byte

	switch is.Encoding {
	case INTSET_ENC_INT16:
		nn := int16(n)
		b = I16ToB(nn)

	case INTSET_ENC_INT32:
		nn := int32(n)
		b = I32ToB(nn)

	case INTSET_ENC_INT64:
		nn := int64(n)
		b = I64ToB(nn)
	}

	for i := 0; i < int(is.Encoding); i++ {
		is.Contents[offset+i] = b[i]
	}
}

// Get returns the integer at given index according to intset's configured encoding.
func (is *IntSet) Get(idx int) (res int) {
	offset := idx * int(is.Encoding)

	switch is.Encoding {
	case INTSET_ENC_INT16:
		return int(BToI16(is.Contents, offset))

	case INTSET_ENC_INT32:
		return int(BToI32(is.Contents, offset))

	case INTSET_ENC_INT64:
		return int(BToI64(is.Contents, offset))

	}
	return
}

// ElementAtIndex returns the integer at given index according to the given encoding.
func (is *IntSet) getEncoded(idx int, enc uint8) (res int) {
	offset := idx * int(enc)

	switch enc {
	case INTSET_ENC_INT16:
		res = int(BToI16(is.Contents, offset))

	case INTSET_ENC_INT32:
		res = int(BToI32(is.Contents, offset))

	case INTSET_ENC_INT64:
		res = int(BToI64(is.Contents, offset))

	}
	fmt.Println("res = ", res)
	return
}

// upgradeAndAdd upgrades the intset to a larger encoding and inserts the given integer.
func (is *IntSet) upgradeAndAdd(n int) {
	currEnc, newEnc := is.Encoding, intsetValueEncoding(n)
	length := int(is.Length)
	var prepend int
	if n < 0 {
		prepend = 1
	}

	// First set new encoding and resize
	is.Encoding = newEnc
	is.resize(length + 1)

	// Upgrade back-to-front so we don't overwrite values.
	// Note that the "prepend" variable is used to make sure we have an empty
	// space at either the beginning or the end of the intset. */

	for i := length-1; i >= 0; i-- {

		// length = 7, resize to 8
		// 0   1   2   3   4   5   6
		// --  --  --  --  --  --  --
		// A   B   C   D   E   F   G
		//
		// prepend
		// 0   1   2   3   4   5   6   7
		// --  --  --  --  --  --  --  --
		// X   A   B   C   D   E   F   G
		//
		// append
		// 0   1   2   3   4   5   6   7
		// --  --  --  --  --  --  --  --
		// A   B   C   D   E   F   G   X
		is.Set(is.getEncoded(i, currEnc), i+prepend)
	}

	if prepend == 1 {
		is.Set(n, 0)
	} else {
		is.Set(n, int(is.Length))
	}

	is.Length++
}

func (is *IntSet) resize(length int) {
	currSize, newSize := len(is.Contents), length*int(is.Encoding)
	is.Contents = append(is.Contents, make([]uint8, newSize-currSize)...)
}

func (is *IntSet) moveTail(idx int) {
	begin, end := idx*int(is.Encoding), int(is.Length)*int(is.Encoding)

	for i := end - 1; i >= begin; i-- {
		is.Contents[i+int(is.Encoding)] = is.Contents[i]
	}
}

func intsetValueEncoding(n int) uint8 {
	if n >= math.MinInt16 && n <= math.MaxInt16 {
		return INTSET_ENC_INT16
	} else if n >= math.MinInt32 && n <= math.MaxInt32 {
		return INTSET_ENC_INT32
	}
	return INTSET_ENC_INT64
}
