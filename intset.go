package main

import (
	"fmt"
	"math"
)

// IntSet is a ordered byte-slice-based data structure that holds integers.
//
// Time Complexity:
// Find		O(logn);
// Add 		O(n);
// Get 		O(n);
type IntSet interface {

	// Find searches for the position of "value".
	// Returns true when the value was found and
	// sets "idx" to the position of the value within the intset.
	// Returns false when the value is not present in the intset
	// and sets "idx" to the position where "value" can be inserted.
	Find(int) (int, bool)

	// Add adds an integer into the intset.
	// The integer wouldn't be inserted if it already exists.
	Add(int) error

	// Get gets the integer at the given index.
	Get(int) (int, error)

	// Length returns the size of the intset.
	Size() int

	// Delete()
}

const (
	IS_ENC_INT16 uint8 = 2
	IS_ENC_INT32 uint8 = 4
	IS_ENC_INT64 uint8 = 8
)

// IntSetImpl has a maximum length of UINT32_MAX
// Padding:
// |Encoding 	|Length |Contents						|
// |XOOO 		|XXXX 	|XXXX XXXX XXXX XXXX XXXX XXXX	|
type IntSetImpl struct {
	Encoding uint8
	Len      uint32
	Contents []uint8
}

func NewIntSet() IntSet {
	is := new(IntSetImpl)
	is.Encoding = IS_ENC_INT16
	return is
}

func (is *IntSetImpl) Size() int {
	return int(is.Len)
}

func (is *IntSetImpl) Find(n int) (int, bool) {
	length := int(is.Len)
	if length == 0 {
		return 0, false
	}

	mi, _ := is.Get(0)
	ma, _ := is.Get(length - 1)
	if n < mi {
		return 0, false
	}
	if n > ma {
		return length, false
	}

	l, r := 0, length-1

	for l < r {
		m := l + (r-l)>>1
		mm, _ := is.Get(m)
		if mm == n {
			return m, true
		} else if mm < n {
			l = m + 1
		} else {
			r = m - 1
		}
	}

	ll, _ := is.Get(l)
	if ll == n {
		return l, true
	} else if ll > n {
		return l, false
	}
	return l + 1, false
}

func (is *IntSetImpl) Add(n int) error {

	if is.Len == math.MaxInt32 {
		return ExceedLimitErr
	}

	// 1. get encoding
	enc := intsetValueEncoding(n)

	// if encoding needs upgrade
	if enc > is.Encoding {
		is.upgradeAndAdd(n)
		return nil
	} else {
		// abort if already in the set
		if idx, exists := is.Find(n); exists {
			return DuplicateInputErr
		} else {
			is.resize(int(is.Len) + 1)
			is.moveTail(idx)
			is.setAtIndex(n, idx)
			is.Len++
			return nil
		}
	}
}

func (is *IntSetImpl) setAtIndex(n, idx int) {
	if idx < 0 || idx > int(is.Len) {
		fmt.Println("invalid input idx")
		return
	}

	offset := idx * int(is.Encoding)
	var b []byte

	switch is.Encoding {
	case IS_ENC_INT16:
		nn := int16(n)
		b = I16ToB(nn)

	case IS_ENC_INT32:
		nn := int32(n)
		b = I32ToB(nn)

	case IS_ENC_INT64:
		nn := int64(n)
		b = I64ToB(nn)
	}

	for i := 0; i < int(is.Encoding); i++ {
		is.Contents[offset+i] = b[i]
	}
}

// Get returns the integer at given index according to intset's configured encoding.
func (is *IntSetImpl) Get(idx int) (int, error) {
	if is.Len == 0 {
		return 0, EmptyErr
	} else if idx < 0 || idx >= int(is.Len) {
		return 0, InvalidIdxErr
	}

	offset := idx * int(is.Encoding)

	switch is.Encoding {
	case IS_ENC_INT16:
		return int(BToI16(is.Contents, offset)), nil

	case IS_ENC_INT32:
		return int(BToI32(is.Contents, offset)), nil

	default:
		return int(BToI64(is.Contents, offset)), nil
	}
}

// ElementAtIndex returns the integer at given index according to the given encoding.
func (is *IntSetImpl) getEncoded(idx int, enc uint8) (res int) {
	offset := idx * int(enc)

	switch enc {
	case IS_ENC_INT16:
		res = int(BToI16(is.Contents, offset))

	case IS_ENC_INT32:
		res = int(BToI32(is.Contents, offset))

	case IS_ENC_INT64:
		res = int(BToI64(is.Contents, offset))

	}
	fmt.Println("res = ", res)
	return
}

// upgradeAndAdd upgrades the intset to a larger encoding and inserts the given integer.
func (is *IntSetImpl) upgradeAndAdd(n int) {
	currEnc, newEnc := is.Encoding, intsetValueEncoding(n)
	length := int(is.Len)
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

	for i := length - 1; i >= 0; i-- {

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
		is.setAtIndex(is.getEncoded(i, currEnc), i+prepend)
	}

	if prepend == 1 {
		is.setAtIndex(n, 0)
	} else {
		is.setAtIndex(n, int(is.Len))
	}

	is.Len++
}

func (is *IntSetImpl) resize(length int) {
	currSize, newSize := len(is.Contents), length*int(is.Encoding)
	is.Contents = append(is.Contents, make([]uint8, newSize-currSize)...)
}

func (is *IntSetImpl) moveTail(idx int) {
	begin, end := idx*int(is.Encoding), int(is.Len)*int(is.Encoding)

	for i := end - 1; i >= begin; i-- {
		is.Contents[i+int(is.Encoding)] = is.Contents[i]
	}
}

func intsetValueEncoding(n int) uint8 {
	if n >= math.MinInt16 && n <= math.MaxInt16 {
		return IS_ENC_INT16
	} else if n >= math.MinInt32 && n <= math.MaxInt32 {
		return IS_ENC_INT32
	}
	return IS_ENC_INT64
}
