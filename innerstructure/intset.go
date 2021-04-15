package innerstructure

import (
	"fmt"
	"math"
	"mykv/util"
)

// IntSet is a ordered byte-slice-based data structure that holds integers.
//
// Time Complexity:
// Find		O(logn);
// Add 		O(n);
// Get 		O(n);
//
// IntSet has a maximum length of UINT32_MAX
// Padding:
// |Encoding 	|Length |Contents						|
// |XOOO 		|XXXX 	|XXXX XXXX XXXX XXXX XXXX XXXX	|
type IntSet struct {
	encoding uint8
	len      uint32
	contents []uint8
}

const (
	IS_ENC_INT16 uint8 = 2
	IS_ENC_INT32 uint8 = 4
	IS_ENC_INT64 uint8 = 8
)

func NewIntSet() *IntSet {
	is := new(IntSet)
	is.encoding = IS_ENC_INT16
	return is
}

// Size returns the size of the intset.
func (is *IntSet) Size() int {
	return int(is.len)
}

// Find searches for the position of "value".
// Returns true when the value was found and
// sets "idx" to the position of the value within the intset.
// Returns false when the value is not present in the intset
// and sets "idx" to the position where "value" can be inserted.
func (is *IntSet) Find(n int) (int, bool) {
	length := int(is.len)
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

// Add adds an integer into the intset.
// The integer wouldn't be inserted if it already exists.
func (is *IntSet) Add(n int) error {

	if is.len == math.MaxInt32 {
		return ErrExceedLimit
	}

	// 1. get encoding
	enc := intsetValueEncoding(n)

	// if encoding needs upgrade
	if enc > is.encoding {
		is.upgradeAndAdd(n)
		return nil
	} else {
		// abort if already in the set
		if idx, exists := is.Find(n); exists {
			return ErrDuplicateInput
		} else {
			is.resize(int(is.len) + 1)
			is.moveTail(idx)
			is.setAtIndex(n, idx)
			is.len++
			return nil
		}
	}
}

// Get returns the integer at given index according to intset's configured encoding.
func (is *IntSet) Get(idx int) (int, error) {
	if is.len == 0 {
		return 0, ErrEmpty
	} else if idx < 0 || idx >= int(is.len) {
		return 0, ErrInvalidIdx
	}

	offset := idx * int(is.encoding)

	switch is.encoding {
	case IS_ENC_INT16:
		return int(util.BToI16(is.contents, offset)), nil

	case IS_ENC_INT32:
		return int(util.BToI32(is.contents, offset)), nil

	default:
		return int(util.BToI64(is.contents, offset)), nil
	}
}

func (is *IntSet) setAtIndex(n, idx int) {
	if idx < 0 || idx > int(is.len) {
		fmt.Println("invalid input idx")
		return
	}

	offset := idx * int(is.encoding)

	switch is.encoding {
	case IS_ENC_INT16:
		nn := int16(n)
		util.I16ToB(nn, is.contents, offset)

	case IS_ENC_INT32:
		nn := int32(n)
		util.I32ToB(nn, is.contents, offset)

	case IS_ENC_INT64:
		nn := int64(n)
		util.I64ToB(nn, is.contents, offset)
	}
}

func (is *IntSet) getEncoded(idx int, enc uint8) (res int) {
	offset := idx * int(enc)

	switch enc {
	case IS_ENC_INT16:
		res = int(util.BToI16(is.contents, offset))

	case IS_ENC_INT32:
		res = int(util.BToI32(is.contents, offset))

	case IS_ENC_INT64:
		res = int(util.BToI64(is.contents, offset))

	}
	fmt.Println("res = ", res)
	return
}

// upgradeAndAdd upgrades the intset to a larger encoding and inserts the given integer.
func (is *IntSet) upgradeAndAdd(n int) {
	currEnc, newEnc := is.encoding, intsetValueEncoding(n)
	length := int(is.len)
	var prepend int
	if n < 0 {
		prepend = 1
	}

	// First set new encoding and resize
	is.encoding = newEnc
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
		is.setAtIndex(n, int(is.len))
	}

	is.len++
}

func (is *IntSet) resize(length int) {
	currSize, newSize := len(is.contents), length*int(is.encoding)
	is.contents = append(is.contents, make([]uint8, newSize-currSize)...)
}

func (is *IntSet) moveTail(idx int) {
	begin, end := idx*int(is.encoding), int(is.len)*int(is.encoding)

	for i := end - 1; i >= begin; i-- {
		is.contents[i+int(is.encoding)] = is.contents[i]
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
