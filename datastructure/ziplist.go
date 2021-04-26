package datastructure

import (
	"bytes"
	"errors"
	"math"
	"github.com/viktorxhzj/mykv/util"
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
// 2. ZipList uses uint16 to store the number of entries,
// so ZipList has at most 1<<16 - 1 = 65535 entries.
type ZipList []byte

const (
	ZL_MAX_LEN = math.MaxUint16

	ZL_TAIL_OFFSET  = 4
	ZL_ZLLEN_OFFSET = 8
	ZL_INIT_SIZE    = 11
	ZL_END          = 0xFF
	ZL_BIG_PREVLEN  = 0xFE

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
	ZL_ENTRY_MAX_SIZE = math.MaxUint32 - 11
)

var (
	ErrZLInvalidInput     = errors.New("input data is neither string or integer")
	ErrZLEntryExceedLimit = errors.New("input data is too large")
)

// Future: supports deletion and cascadeUpdate

func NewZipList() *ZipList {
	z := new(ZipList)
	*z = append(*z, make([]byte, 11)...)
	util.UI32ToB(ZL_INIT_SIZE, *z, 0)
	util.UI32ToB(ZL_INIT_SIZE-1, *z, ZL_TAIL_OFFSET)
	util.UI16ToB(0, *z, ZL_ZLLEN_OFFSET)
	util.UI8ToB(ZL_END, *z, ZL_INIT_SIZE-1)
	return z
}

// ZLBytes returns the number of bytes that the ziplist occupies.
func (z *ZipList) ZLBytes() int {
	return int(util.BToUI32(*z, 0))
}

// ZLTail returns the offset to the last entry in the list.
func (z *ZipList) ZLTail() int {
	return int(util.BToUI32(*z, ZL_TAIL_OFFSET))
}

// ZLLen returns the number of entries.
func (z *ZipList) ZLLen() int {
	return int(util.BToUI16(*z, ZL_ZLLEN_OFFSET))
}

// Add adds an integer to the ziplist.
func (z *ZipList) AddInt(i int) error {
	if z.ZLLen() == ZL_MAX_LEN {
		return ErrExceedLimit
	}
	return z.zipListInsert(i)
}

// Add adds a string to the ziplist.
func (z *ZipList) AddString(s string) error {
	if z.ZLLen() == ZL_MAX_LEN {
		return ErrExceedLimit
	}
	return z.zipListInsert([]byte(s))
}

// Get returns the element at the given index of the ziplist.
func (z *ZipList) Get(idx int) (interface{}, error) {
	if l := z.ZLLen(); l == 0 {
		return nil, ErrEmpty
	} else if idx < 0 || idx >= l {
		return nil, ErrInvalidIdx
	} else {
		p := z.ZLTail()
		for i := 0; i < l-idx-1; i++ {
			e := z.newZipListEntry(p)
			p -= e.PrevRawLen
		}
		e := z.newZipListEntry(p)
		if e.Encoding < ZL_STR_MASK {
			return string(z.loadString(p+int(e.HeaderSize), e.Len)), nil
		} else if e.Encoding >= ZL_INT_IMM_MIN && e.Encoding <= ZL_INT_IMM_MAX {
			return int(e.Encoding - ZL_INT_IMM_MIN), nil
		} else {
			return z.loadInteger(p+int(e.HeaderSize), e.Len), nil
		}
	}
}

// FindInt searches for the position of the given integer in the ziplist and returns the index.
// If the ziplist does not contain the given element, it returns -1.
func (z *ZipList) FindInt(ii int) int {
	if l := z.ZLLen(); l == 0 {
		return -1
	} else {
		p := z.ZLTail()
		for i := 0; i < l; i++ {
			e := z.newZipListEntry(p)
			if e.Encoding >= ZL_INT_IMM_MIN && e.Encoding <= ZL_INT_IMM_MAX && int(e.Encoding-ZL_INT_IMM_MIN) == ii {
				return l - i - 1
			} else if e.Encoding >= ZL_STR_MASK && z.loadInteger(p+int(e.HeaderSize), e.Len) == ii {
				return l - i - 1
			}
			p -= e.PrevRawLen
		}
	}
	return -1
}

// FindInt searches for the position of the given string in the ziplist and returns the index.
// If the ziplist does not contain the given element, it returns -1.
func (z *ZipList) FindString(ss string) int {
	if l := z.ZLLen(); l == 0 {
		return -1
	} else {
		p := z.ZLTail()
		for i := 0; i < l; i++ {
			e := z.newZipListEntry(p)
			if e.Encoding < ZL_STR_MASK && bytes.Equal(z.loadString(p+int(e.HeaderSize), e.Len), []byte(ss)) {
				return l - i - 1
			}
			p -= e.PrevRawLen
		}
	}
	return -1
}

func (z *ZipList) updateZLTail(tailPos int) {
	util.UI32ToB(uint32(tailPos), *z, ZL_TAIL_OFFSET)
}

func (z *ZipList) updateZLLen() {
	util.UI16ToB(uint16(z.ZLLen() + 1), *z, ZL_ZLLEN_OFFSET)
}

func (z *ZipList) updateZLBytes() {
	util.UI32ToB(uint32(len(*z)), *z, 0)
}

func (z *ZipList) storePrevEntryLength(p, prevLen int) (reqLen int) {
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
		util.UI32ToB(uint32(prevLen), *z, p+1)
	}
	return
}

func (z *ZipList) storeEntryStringEncoding(p int, s []byte) (encoding uint8, reqLen int) {
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
		util.UI32ToB(uint32(rawLen), *z, p+1)
	}
	return
}

func (z *ZipList) storeEntryIntegerEncoding(p int, n int) (encoding uint8, reqLen int) {
	reqLen = 1

	switch {
	case n >= 0 && n <= ZL_INT_IMM_MAX-ZL_INT_IMM_MIN:
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

func (z *ZipList) storeInteger(p, n int) {
	switch {
	case n >= 0 && n <= ZL_INT_IMM_MAX-ZL_INT_IMM_MIN:

	case n >= math.MinInt8 && n <= math.MaxInt8:
		util.I8ToB(int8(n), *z, p)
	case n >= math.MinInt16 && n <= math.MaxInt16:
		util.I16ToB(int16(n), *z, p)
	case n >= math.MinInt32 && n <= math.MaxInt32:
		util.I32ToB(int32(n), *z, p)
	default:
		util.I64ToB(int64(n), *z, p)
	}
}

func (z *ZipList) loadInteger(p, len int) int {
	switch len {
	case 0:
		return 0

	case 1:
		return int(util.BToI8(*z, p))
	case 2:
		return int(util.BToI16(*z, p))
	case 4:
		return int(util.BToI32(*z, p))
	default:
		return int(util.BToI64(*z, p))
	}
}

func (z *ZipList) storeString(p int, s []byte) {
	for i := 0; i < len(s); i++ {
		(*z)[p+i] = s[i]
	}
}

func (z *ZipList) loadString(p, len int) []byte {
	return (*z)[p : p+len]
}

// zipListInsert inserts the element at the tail.
// This is different from Redis implementation as Redis supports insertion at given index other than the tail.
func (z *ZipList) zipListInsert(e interface{}) error {
	var curLen, reqLen, newLen, prevLen int

	curLen = len(*z)

	// insertion position and tail position
	p, pt := len(*z)-1, z.ZLTail()

	// if the ziplist is not empty
	if (*z)[pt] != ZL_END {
		pte := z.newZipListEntry(pt)
		prevLen = pte.Len + int(pte.HeaderSize)
	}

	// calculate required length for this entry, and determine the encoding byte
	s, ok1 := e.([]byte)
	if ok1 {
		_, add := z.storeEntryStringEncoding(-1, s)
		reqLen += len(s)
		reqLen += add
	}
	i, ok2 := e.(int)
	if ok2 {
		encoding, add := z.storeEntryIntegerEncoding(-1, i)
		reqLen += intSizeByEncoding(encoding)
		reqLen += add
	}

	add1 := z.storePrevEntryLength(-1, prevLen)
	reqLen += add1

	if reqLen > ZL_ENTRY_MAX_SIZE {
		return ErrZLEntryExceedLimit
	}

	// resize and update ZLBytes, ZLTail
	newLen = curLen + reqLen
	*z = append(*z, make([]byte, reqLen)...)
	(*z)[newLen-1] = ZL_END
	z.updateZLBytes()
	z.updateZLTail(p)

	z.storePrevEntryLength(p, prevLen)
	if ok1 {
		_, add2 := z.storeEntryStringEncoding(p+add1, s)
		z.storeString(p+add1+add2, s)
	}
	if ok2 {
		_, add2 := z.storeEntryIntegerEncoding(p+add1, i)
		z.storeInteger(p+add1+add2, i)
	}

	// update ZLLen
	z.updateZLLen()
	return nil
}

// decodePrevLen returns prevLen and the size it takes (1 or 5).
func (z *ZipList) decodePrevLen(p int) (prevLenSize uint8, prevLen int) {
	if (*z)[p] < ZL_BIG_PREVLEN {
		prevLenSize = 1
		prevLen = int(util.BToUI8(*z, p))
	} else {
		prevLenSize = 5
		prevLen = int(util.BToUI32(*z, p+1))
	}
	return
}

func (z *ZipList) decodeEncoding(p int) (e byte) {
	e = (*z)[p]
	if e < ZL_STR_MASK {
		e &= ZL_STR_MASK
	}
	return
}

func (z *ZipList) decodeLen(p int, encoding uint8) (lenSize uint8, len int) {
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
			len = int(util.BToUI32(*z, p+1))
		}
	} else {
		lenSize = 1
		len = intSizeByEncoding(encoding)
	}
	return
}

type zipListEntry struct {
	PrevRawLenSize uint8 // Bytes used to encode the previous entry len
	LenSize        uint8 // Bytes used to encode this entry
	HeaderSize     uint8 // PrevRawLenSize + LenSize
	Encoding       uint8 // Set to ZIP_STR_* or ZIP_INT_* depending on the encoding
	PrevRawLen     int   // Previous entry len
	Len            int   // Bytes used to represent the actual entry
	Offset         int   // Pointer to the very start of the entry

	// total 32 bytes
}

func (z *ZipList) newZipListEntry(p int) (e zipListEntry) {
	e.PrevRawLenSize, e.PrevRawLen = z.decodePrevLen(p)
	e.Encoding = z.decodeEncoding(p + int(e.PrevRawLenSize))
	e.LenSize, e.Len = z.decodeLen(p+int(e.PrevRawLenSize), e.Encoding)
	e.HeaderSize = e.PrevRawLenSize + e.LenSize
	e.Offset = p
	return e
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
