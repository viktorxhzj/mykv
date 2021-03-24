package main

import (
	"bytes"
	"fmt"
	"math"
)

// ZipList is a byte-slice-based data structure
// that holds integers, strings, or byte slices.
//
// Time Complexity:
// Add 		O(1);
// Get 		O(n);
// Find 	O(mn);
//
// Size:
// 1. ZipList uses uint32 to store the number of bytes it occupies,
// so ZipList is at most 1<<32 -1 bytes = 4096 MB.
//
// 2. ZipList uses uint16 to store the number of entries,
// so ZipList has at most 1<<16 - 1 = 65535 entries.
type ZipList interface {

	// ZLBytes returns the number of bytes that the ziplist occupies.
	ZLBytes() int

	// ZLTail returns the offset to the last entry in the list.
	ZLTail() int

	// ZLLen returns the number of entries.
	ZLLen() int
	
	// Find searches for the position of the given element in the ziplist and returns the index.
	// If the ziplist does not contain the given element, it returns -1.
	Find(interface{}) int

	// Add adds an element to the ziplist. If the given element
	// is either an integer or a string, it returns false.
	Add(interface{}) bool

	// Get returns the element at the given index of the ziplist.
	// If the ziplist is empty or the index is beyond the size, returns nil.
	Get(int) interface{}
}

const (
	ZIPLIST_MAX_LENGTH = math.MaxUint16

	ZIPLIST_ZLTAIL_OFFSET = 4
	ZIPLIST_ZLLEN_OFFSET  = 8
	ZIPLIST_INIT_SIZE     = 11
	ZIPLIST_ZLEND         = 0xFF

	ZIPLISTENTRY_1BYTE_MAX_SIZE = 1<<6 - 1

	ZIPLISTENTRY_2BYTES_MAX_SIZE = 1<<14 - 1

	// An empty ZipList takes 11 bytes. Therefore,
	// a ZipListEntry can take math.MaxUint32-11 atmost.
	ZIPLISTENTRY_5BYTES_MAX_SIZE = math.MaxUint32 - 11
)

// <zlbytes> <zltail> <zllen> <entry> <entry> ... <entry> <zlend>
// uint32    uint32   uint16							  uint8
// ZipListImpl can store a string that takes ZIPLISTENTRY_5BYTES_MAX_SIZE bits at most, or an int64 value.
type ZipListImpl []byte

func (z *ZipListImpl) Append(b ...byte) {
	for _, v := range(b) {
		*z = append(*z, v)
	}
}

func NewZipList() ZipList {
	z := new(ZipListImpl)
	z.Append(UI32ToB(ZIPLIST_INIT_SIZE)...)
	z.Append(UI32ToB(0)...)
	z.Append(UI16ToB(0)...)
	z.Append(UI8ToB(ZIPLIST_ZLEND)...)
	return z
}

func (z *ZipListImpl) ZLBytes() int {
	return len(*z)
}

func (z *ZipListImpl) ZLTail() int {
	return int(BToUI32(*z, ZIPLIST_ZLTAIL_OFFSET))
}

func (z *ZipListImpl) ZLLen() int {
	return int(BToUI16(*z, ZIPLIST_ZLLEN_OFFSET))
}

func (z *ZipListImpl) Add(data interface{}) bool {
	s, ok1 := data.(string)
	b, ok2 := data.([]byte)
	i, ok3 := data.(int)

	if z.ZLLen() == ZIPLIST_MAX_LENGTH {
		fmt.Println("zip list reaches the maximum size")
		return false
	}

	if !ok1 && !ok2 && !ok3 {
		fmt.Println("input data is nonf of integer, string or byte slice")
		return false
	}

	var prevLen, encoding, entry []byte

	if ok1 || ok2 {
		if ok1 && len([]byte(s)) >= ZIPLISTENTRY_5BYTES_MAX_SIZE {
			fmt.Println("input string is too long")
			return false
		}
		if ok2 && len(b) >= ZIPLISTENTRY_5BYTES_MAX_SIZE {
			fmt.Println("input byte slice is too long")
			return false
		}
		if ok1 {
			entry = []byte(s)
		} else {
			entry = b
		}
		encoding = z.encodeBytes(entry)
	}

	if ok3 {
		encoding, entry = z.encodeInteger(i)
	}

	prevLen = z.calculatePrevLen()

	z.updateZLTail()

	*z = (*z)[:len(*z)-1]
	z.Append(prevLen...)
	z.Append(encoding...)
	z.Append(entry...)
	z.Append(UI8ToB(ZIPLIST_ZLEND)...)

	z.updateZLLen()

	return true
}

func (z *ZipListImpl) Get(idx int) (res interface{}) {
	if l := z.ZLLen(); l == 0 {
		return
	} else if idx >= l {
		return
	} else {
		c := z.ZLTail()

		var prevLen, currLen uint32
		currLen = uint32(z.ZLBytes()-c) - 1

		for i := 0; i < l-idx; i++ {
			var offset int

			if (*z)[c] != 0xFE {
				prevLen = uint32((*z)[c])
				offset += 1
			} else {
				prevLen = BToUI32(*z, int(c+1))
				offset += 5
			}

			if i == l-idx-1 {

				if (*z)[c+offset]>>6 <= 2 {
					return z.decodeBytes(c, offset, currLen)
				} else {
					return z.decodeInteger(c, offset, currLen)
				}
			}

			c -= int(prevLen)
			currLen = prevLen
		}
	}
	return
}

func (z *ZipListImpl) Find(data interface{}) int {
	ss, ok1 := data.(string)
	bb, ok2 := data.([]byte)
	ii, ok3 := data.(int)

	if !ok1 && !ok2 && !ok3 {
		fmt.Println("input data is nonf of integer, string or byte slice")
		return -1
	}

	if l := z.ZLLen(); l == 0 {
		return -1
	} else {
		c := z.ZLTail()

		var prevLen, currLen uint32
		currLen = uint32(z.ZLBytes()-c) - 1

		for i := 0; i < l; i++ {
			var offset int

			if (*z)[c] != 0xFE {
				prevLen = uint32((*z)[c])
				offset += 1
			} else {
				prevLen = BToUI32(*z, int(c+1))
				offset += 5
			}

			// if the stored element is a byte slice and the input is a string
			if (*z)[c+offset]>>6 <= 2 && ok1 && bytes.Equal([]byte(ss), z.decodeBytes(c, offset, currLen)) {
				return int(l - i - 1)
			}

			// if the stored element is a byte slice and the input is a byte slice
			if (*z)[c+offset]>>6 <= 2 && ok2 && bytes.Equal(bb, z.decodeBytes(c, offset, currLen)) {
				return int(l - i - 1)
			}

			// if the stored element is an integer and input is an integer
			if (*z)[c+offset]>>6 > 2 && ok3 && ii == z.decodeInteger(c, offset, currLen) {
				return int(l - i - 1)
			}

			c -= int(prevLen)
			currLen = prevLen
		}
	}
	return -1
}

func (z *ZipListImpl) updateZLTail() {
	b := UI32ToB(uint32(z.ZLBytes() - 1))
	for i := 0; i < 4; i++ {
		(*z)[ZIPLIST_ZLTAIL_OFFSET+i] = b[i]
	}
}

func (z *ZipListImpl) updateZLLen() {
	b := UI16ToB(uint16(z.ZLLen() + 1))
	for i := 0; i < 2; i++ {
		(*z)[ZIPLIST_ZLLEN_OFFSET+i] = b[i]
	}
}

func (z *ZipListImpl) encodeBytes(entry []byte) (encoding []byte) {
	l := len(entry)
	switch {
	case l <= ZIPLISTENTRY_1BYTE_MAX_SIZE:
		encoding = UI8ToB(uint8(l))
	case l <= ZIPLISTENTRY_2BYTES_MAX_SIZE:
		encoding = UI16ToB(uint16(l) | 0x4000)
	case l <= ZIPLISTENTRY_5BYTES_MAX_SIZE:
		encoding = UI8ToB(0x80)
		encoding = append(encoding, UI32ToB(uint32(l))...)
	}
	return
}

func (z *ZipListImpl) decodeBytes(c, offset int, currLen uint32) []byte {
	switch (*z)[c+offset] >> 6 {
	case 0x00:
		offset += 1
	case 0x01:
		offset += 2
	case 0x02:
		offset += 5
	}

	return (*z)[c+offset : c+int(currLen)]
}

func (z *ZipListImpl) encodeInteger(i int) (encoding, entry []byte) {
	switch {
	case i >= 0 && i <= 13:
		encoding = UI8ToB(0xF0 | uint8(i))

	case i >= math.MinInt8 && i <= math.MaxInt8:
		encoding = UI8ToB(0xFE)
		entry = I8ToB(int8(i))

	case i >= math.MinInt16 && i <= math.MaxInt16:
		encoding = UI8ToB(0xC0)
		entry = I16ToB(int16(i))

	case i >= math.MinInt32 && i <= math.MaxInt32:
		encoding = UI8ToB(0xD0)
		entry = I32ToB(int32(i))

	default:
		encoding = UI8ToB(0xE0)
		entry = I64ToB(int64(i))
	}
	return
}

func (z *ZipListImpl) decodeInteger(c, offset int, currLen uint32) int {
	var i int
	switch (*z)[c+offset] {
	case 0xFE:
		offset += 1
		i = int(int8((*z)[c+offset]))

	case 0xC0:
		offset += 1
		i = int(BToI16((*z), int(c+offset)))

	case 0xD0:
		offset += 1
		i = int(BToI32((*z), int(c+offset)))

	case 0xE0:
		offset += 1
		i = int(BToI64((*z), int(c+offset)))

	default:
		i = int(uint8((*z)[c+offset] & 0x0F))
	}
	return i
}

func (z *ZipListImpl) calculatePrevLen() (prevLen []byte) {

	tailOffset := z.ZLTail()

	if tailOffset == 0 {
		prevLen = make([]byte, 1)
	} else {
		p := z.ZLBytes() - 1 - tailOffset
		if p < 0xFE {
			prevLen = UI8ToB(uint8(p))
		} else {
			prevLen = UI8ToB(0xFE)
			prevLen = append(prevLen, UI32ToB(uint32(z.ZLBytes()-1-tailOffset))...)
		}
	}
	return
}


func (z *ZipListImpl) zipListInsert() {

}