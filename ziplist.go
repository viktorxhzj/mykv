package main

import (
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
	ZL_END                = 0xFF
	ZL_BIG_PREVLEN        = 0xFE

	ZL_STR_MASK = 0xC0
	ZL_STR_06B  = 0 << 6
	ZL_STR_14B  = 1 << 6
	ZL_STR_32B  = 2 << 6

	ZL_INT_16B = 0xC0 | 0<<4
	ZL_INT_32B = 0xC0 | 1<<4
	ZL_INT_64B = 0xC0 | 2<<4
	ZL_INT_8B  = 0xFE

	ZL_INT_IMM_MIN = 0xF0
	ZL_INT_IMM_MAX = 0xFD

	ZL_STR_06B_MAX_SIZE = 1<<6 - 1

	ZL_STR_14B_MAX_SIZE = 1<<14 - 1

	// An empty ZipList takes 11 bytes. Therefore,
	// a ZipListEntry can take math.MaxUint32-11 atmost.
	ZIPLISTENTRY_5BYTES_MAX_SIZE = math.MaxUint32 - 11
)

// <zlbytes> <zltail> <zllen> <entry> <entry> ... <entry> <zlend>
// uint32    uint32   uint16							  uint8
type ZipListImpl []byte

func (z *ZipListImpl) Append(b ...byte) {
	for _, v := range b {
		*z = append(*z, v)
	}
}

// func NewZipList() ZipList {
// 	z := new(ZipListImpl)
// 	z.Append(UI32ToB(ZIPLIST_INIT_SIZE)...)
// 	z.Append(UI32ToB(ZIPLIST_INIT_SIZE - 1)...)
// 	z.Append(UI16ToB(0)...)
// 	z.Append(UI8ToB(ZL_END)...)
// 	return z
// }

func (z *ZipListImpl) ZLBytes() int {
	return len(*z)
}

func (z *ZipListImpl) ZLTail() int {
	return int(BToUI32(*z, ZIPLIST_ZLTAIL_OFFSET))
}

func (z *ZipListImpl) ZLLen() int {
	return int(BToUI16(*z, ZIPLIST_ZLLEN_OFFSET))
}

// func (z *ZipListImpl) Add(data interface{}) bool {
// 	s, ok1 := data.(string)
// 	b, ok2 := data.([]byte)
// 	i, ok3 := data.(int)

// 	if z.ZLLen() == ZIPLIST_MAX_LENGTH {
// 		fmt.Println("zip list reaches the maximum size")
// 		return false
// 	}

// 	if !ok1 && !ok2 && !ok3 {
// 		fmt.Println("input data is nonf of integer, string or byte slice")
// 		return false
// 	}

// 	var prevLen, encoding, entry []byte

// 	if ok3 {
// 		encoding, entry = z.encodeInteger(i)
// 	}

// 	prevLen = z.calculatePrevLen()

// 	z.updateZLTail()

// 	*z = (*z)[:len(*z)-1]
// 	z.Append(prevLen...)
// 	z.Append(encoding...)
// 	z.Append(entry...)
// 	z.Append(UI8ToB(ZL_END)...)

// 	z.updateZLLen()

// 	return true
// }

// func (z *ZipListImpl) Get(idx int) (res interface{}) {
// 	if l := z.ZLLen(); l == 0 {
// 		return
// 	} else if idx >= l {
// 		return
// 	} else {
// 		c := z.ZLTail()

// 		var prevLen, currLen uint32
// 		currLen = uint32(z.ZLBytes()-c) - 1

// 		for i := 0; i < l-idx; i++ {
// 			var offset int

// 			if (*z)[c] != 0xFE {
// 				prevLen = uint32((*z)[c])
// 				offset += 1
// 			} else {
// 				prevLen = BToUI32(*z, int(c+1))
// 				offset += 5
// 			}

// 			if i == l-idx-1 {

// 				if (*z)[c+offset]>>6 <= 2 {
// 					return z.decodeBytes(c, offset, currLen)
// 				} else {
// 					return z.decodeInteger(c, offset, currLen)
// 				}
// 			}

// 			c -= int(prevLen)
// 			currLen = prevLen
// 		}
// 	}
// 	return
// }

// func (z *ZipListImpl) Find(data interface{}) int {
// 	ss, ok1 := data.(string)
// 	bb, ok2 := data.([]byte)
// 	ii, ok3 := data.(int)

// 	if !ok1 && !ok2 && !ok3 {
// 		fmt.Println("input data is nonf of integer, string or byte slice")
// 		return -1
// 	}

// 	if l := z.ZLLen(); l == 0 {
// 		return -1
// 	} else {
// 		c := z.ZLTail()

// 		var prevLen, currLen uint32
// 		currLen = uint32(z.ZLBytes()-c) - 1

// 		for i := 0; i < l; i++ {
// 			var offset int

// 			if (*z)[c] != 0xFE {
// 				prevLen = uint32((*z)[c])
// 				offset += 1
// 			} else {
// 				prevLen = BToUI32(*z, int(c+1))
// 				offset += 5
// 			}

// 			// if the stored element is a byte slice and the input is a string
// 			if (*z)[c+offset]>>6 <= 2 && ok1 && bytes.Equal([]byte(ss), z.decodeBytes(c, offset, currLen)) {
// 				return int(l - i - 1)
// 			}

// 			// if the stored element is a byte slice and the input is a byte slice
// 			if (*z)[c+offset]>>6 <= 2 && ok2 && bytes.Equal(bb, z.decodeBytes(c, offset, currLen)) {
// 				return int(l - i - 1)
// 			}

// 			// if the stored element is an integer and input is an integer
// 			if (*z)[c+offset]>>6 > 2 && ok3 && ii == z.decodeInteger(c, offset, currLen) {
// 				return int(l - i - 1)
// 			}

// 			c -= int(prevLen)
// 			currLen = prevLen
// 		}
// 	}
// 	return -1
// }

// func (z *ZipListImpl) updateZLTail() {
// 	b := UI32ToB(uint32(z.ZLBytes() - 1))
// 	for i := 0; i < 4; i++ {
// 		(*z)[ZIPLIST_ZLTAIL_OFFSET+i] = b[i]
// 	}
// }

// func (z *ZipListImpl) updateZLLen() {
// 	b := UI16ToB(uint16(z.ZLLen() + 1))
// 	for i := 0; i < 2; i++ {
// 		(*z)[ZIPLIST_ZLLEN_OFFSET+i] = b[i]
// 	}
// }

// 	return (*z)[c+offset : c+int(currLen)]
// }

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

func intSizeByEncoding(encoding uint8) (len int) {
	switch encoding {
	case ZL_INT_8B:
		len = 1
	case ZL_INT_16B:
		len = 2
	case ZL_INT_32B:
		len = 4
	case ZL_INT_64B:
		len = 8
	default:
		len = 0
	}
	return
}

func tryEncoding(n int) (encoding uint8) {
	switch {
	case n >= 0 && n <= 13:
		encoding = ZL_INT_IMM_MIN + uint8(n)

	case n >= math.MinInt8 && n <= math.MaxInt8:
		encoding = ZL_INT_8B

	case n >= math.MinInt16 && n <= math.MaxInt16:
		encoding = ZL_INT_16B

	case n >= math.MinInt32 && n <= math.MaxInt32:
		encoding = ZL_INT_32B

	default:
		encoding = ZL_INT_64B
	}
	return
}

func (z *ZipListImpl) storePrevEntryLength(p, prevLen int) (reqLen int) {
	if prevLen < ZL_BIG_PREVLEN {
		reqLen = 1
		if p == -1 {
			return
		}
		(*z)[p] = uint8(prevLen)
		return
	} else {
		reqLen = 5
		if p == -1 {
			return
		}
		(*z)[p] = ZL_BIG_PREVLEN
		b := UI32ToB(uint32(prevLen))
		for i := 0; i < 4; i++ {
			(*z)[p+i+1] = b[i]
		}
	}
	return
}

func (z *ZipListImpl) storeEntryStringEncoding(p int, s []byte) (encoding uint8, reqLen int) {
	reqLen = 1
	rawLen := len(s)
	if rawLen <= ZL_STR_06B_MAX_SIZE {
		encoding = ZL_STR_06B
		if p == -1 {
			return
		}
		(*z)[p] = uint8(rawLen)
	} else if rawLen <= ZL_STR_14B_MAX_SIZE {
		encoding = ZL_STR_14B
		reqLen += 1
		if p == -1 {
			return
		}
		(*z)[p] = ZL_STR_14B | (uint8(rawLen>>8) & 0x3F)
		(*z)[p+1] = uint8(rawLen) & 0xFF
	} else {
		encoding = ZL_STR_32B
		reqLen += 4
		if p == -1 {
			return
		}
		(*z)[p] = ZL_STR_32B
		(*z)[p+1] = uint8(rawLen>>24) & 0xFF
		(*z)[p+2] = uint8(rawLen>>16) & 0xFF
		(*z)[p+3] = uint8(rawLen>>8) & 0xFF
		(*z)[p+4] = uint8(rawLen) & 0xFF
	}
	return
}

func (z *ZipListImpl) storeEntryIntegerEncoding(p int, n int) (encoding uint8, reqLen int) {
	reqLen = 1

	switch {
	case n >= 0 && n <= 13:
		encoding = ZL_INT_IMM_MIN + uint8(n)

	case n >= math.MinInt8 && n <= math.MaxInt8:
		encoding = ZL_INT_8B

	case n >= math.MinInt16 && n <= math.MaxInt16:
		encoding = ZL_INT_16B

	case n >= math.MinInt32 && n <= math.MaxInt32:
		encoding = ZL_INT_32B

	default:
		encoding = ZL_INT_64B
	}

	if p != -1 {
		(*z)[p] = encoding
	}

	return
}

// zipListInsert inserts the element at the given position.
// If the element is inserted at the tail, z[p] = ZL_END.
func (z *ZipListImpl) zipListInsert(e interface{}, p int) {
	var prevLenSize, encoding uint8
	var curLen, reqLen, newLen, prevLen int
	var offset, nextDiff, value int

	curLen = len(*z)

	if (*z)[p] != ZL_END {
		prevLenSize, prevLen = z.decodePrevLen(p)
	} else {
		pTail := z.ZLTail()
		if (*z)[pTail] != ZL_END {
			prevLen = z.newZipListEntry(pTail).PrevRawLen
		}
	}

	if s, ok := e.([]byte); ok {
		reqLen = len(s)
		encoding, add := z.storeEntryStringEncoding(-1, s)
		reqLen += add
	} else if i, ok := e.(int); ok {
		encoding, add := z.storeEntryIntegerEncoding(-1, i)
		reqLen = intSizeByEncoding(encoding)
		reqLen += add
	}

	reqLen += z.storePrevEntryLength(-1, prevLen)

}

// decodePrevLen returns prevLen and the size it takes (1 or 5).
func (z *ZipListImpl) decodePrevLen(p int) (prevLenSize uint8, prevLen int) {
	if (*z)[p] < ZL_BIG_PREVLEN {
		prevLenSize = 1
		prevLen = int(BToUI8(*z, p))
	} else {
		prevLenSize = 5
		prevLen = int(BToUI32(*z, p+1))
	}
	return
}

func (z *ZipListImpl) decodeEncoding(p int) (e byte) {
	e = (*z)[p]
	if e < ZL_STR_MASK {
		e &= ZL_STR_MASK
	}
	return
}

func (z *ZipListImpl) decodeLen(p int, encoding uint8) (lenSize uint8, len int) {
	if encoding < ZL_STR_MASK {
		switch encoding {
		case ZL_STR_06B:
			lenSize = 1
			len = int((*z)[p])
		case ZL_STR_14B:
			lenSize = 2
			len = int((uint16((*z)[p]&0x3F) << 8) | uint16((*z)[p+1]))
		case ZL_STR_32B:
			lenSize = 5
			len = int(BToUI32(*z, p+1))
		}
	} else {
		lenSize = 1
		len = intSizeByEncoding(encoding)
	}
	return
}

type zipListEntry struct {
	PrevRawLenSize uint8 // 1 Bytes used to encode the previous entry len
	LenSize        uint8 // 1 Bytes used to encode this entry
	HeaderSize     uint8 // 1 PrevRawLenSize + LenSize
	Encoding       uint8 // 1 Set to ZIP_STR_* or ZIP_INT_* depending on the encoding
	PrevRawLen     int   // 8 Previous entry len
	Len            int   // 8 Bytes used to represent the actual entry
	Offset         int   // 8 Pointer to the very start of the entry

	// total 32 bytes
}

func (z *ZipListImpl) newZipListEntry(p int) (e zipListEntry) {
	e.PrevRawLenSize, e.PrevRawLen = z.decodePrevLen(p)
	e.Encoding = z.decodeEncoding(p + int(e.PrevRawLenSize))
	e.LenSize, e.Len = z.decodeLen(p+int(e.PrevRawLenSize), e.Encoding)
	e.HeaderSize = e.PrevRawLenSize + e.LenSize
	e.Offset = p
	return e
}
