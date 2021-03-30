package innerstructure

import (
	"bytes"
	"errors"
	"math"
	"mykv/util"
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

	// Add adds an element to the ziplist.
	Add(interface{}) error

	// Get returns the element at the given index of the ziplist.
	Get(int) (interface{}, error)

	// Delete(int)
}

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
	ZLInvalidInputErr     = errors.New("input data is neither string or integer")
	ZLEntryExceedLimitErr = errors.New("input data is too large")
)

// <zlbytes> <zltail> <zllen> <entry> <entry> ... <entry> <zlend>
// uint32    uint32   uint16							  uint8
type ZipListImpl []byte

// Future: supports deletion and cascadeUpdate

func (z *ZipListImpl) Append(b ...byte) {
	for _, v := range b {
		*z = append(*z, v)
	}
}

func NewZipList() ZipList {
	z := new(ZipListImpl)
	z.Append(util.UI32ToB(ZL_INIT_SIZE)...)
	z.Append(util.UI32ToB(ZL_INIT_SIZE - 1)...)
	z.Append(util.UI16ToB(0)...)
	z.Append(util.UI8ToB(ZL_END)...)
	return z
}

func (z *ZipListImpl) ZLBytes() int {
	return int(util.BToUI32(*z, 0))
}

func (z *ZipListImpl) ZLTail() int {
	return int(util.BToUI32(*z, ZL_TAIL_OFFSET))
}

func (z *ZipListImpl) ZLLen() int {
	return int(util.BToUI16(*z, ZL_ZLLEN_OFFSET))
}

func (z *ZipListImpl) Add(e interface{}) error {
	if z.ZLLen() == ZL_MAX_LEN {
		return ExceedLimitErr
	}

	s, i, t := util.AssertValidType(e)

	switch t {
	case 0:
		return z.zipListInsert(s)
	case 1:
		return z.zipListInsert(i)
	default:
		return ZLInvalidInputErr
	}
}

func (z *ZipListImpl) Get(idx int) (interface{}, error) {
	if l := z.ZLLen(); l == 0 {
		return nil, EmptyErr
	} else if idx < 0 || idx >= l {
		return nil, InvalidIdxErr
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

func (z *ZipListImpl) Find(e interface{}) int {
	ss, ii, t := util.AssertValidType(e)

	if t == -1 {
		return -1
	}

	if l := z.ZLLen(); l == 0 {
		return -1
	} else {
		p := z.ZLTail()
		for i := 0; i < l; i++ {
			e := z.newZipListEntry(p)
			if t == 0 && e.Encoding < ZL_STR_MASK && bytes.Equal(z.loadString(p+int(e.HeaderSize), e.Len), ss) {
				return l - i - 1
			} else if t == 1 && e.Encoding >= ZL_INT_IMM_MIN && e.Encoding <= ZL_INT_IMM_MAX && int(e.Encoding-ZL_INT_IMM_MIN) == ii {
				return l - i - 1
			} else if t == 1 && e.Encoding >= ZL_STR_MASK && z.loadInteger(p+int(e.HeaderSize), e.Len) == ii {
				return l - i - 1
			}
			p -= e.PrevRawLen
		}
	}
	return -1
}

func (z *ZipListImpl) updateZLTail(tailPos int) {
	b := util.UI32ToB(uint32(tailPos))
	for i := 0; i < 4; i++ {
		(*z)[ZL_TAIL_OFFSET+i] = b[i]
	}
}

func (z *ZipListImpl) updateZLLen() {
	b := util.UI16ToB(uint16(z.ZLLen() + 1))
	for i := 0; i < 2; i++ {
		(*z)[ZL_ZLLEN_OFFSET+i] = b[i]
	}
}

func (z *ZipListImpl) updateZLBytes() {
	b := util.UI32ToB(uint32(len(*z)))
	for i := 0; i < 4; i++ {
		(*z)[i] = b[i]
	}
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
		b := util.UI32ToB(uint32(prevLen))
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
		b := util.UI32ToB(uint32(rawLen))
		for i := 0; i < 4; i++ {
			(*z)[p+i+1] = b[i]
		}
	}
	return
}

func (z *ZipListImpl) storeEntryIntegerEncoding(p int, n int) (encoding uint8, reqLen int) {
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

func (z *ZipListImpl) storeInteger(p, n int) {
	switch {
	case n >= 0 && n <= ZL_INT_IMM_MAX-ZL_INT_IMM_MIN:

	case n >= math.MinInt8 && n <= math.MaxInt8:
		(*z)[p] = util.I8ToB(int8(n))[0]
	case n >= math.MinInt16 && n <= math.MaxInt16:
		b := util.I16ToB(int16(n))
		for i := 0; i < 2; i++ {
			(*z)[p+i] = b[i]
		}
	case n >= math.MinInt32 && n <= math.MaxInt32:
		b := util.I32ToB(int32(n))
		for i := 0; i < 4; i++ {
			(*z)[p+i] = b[i]
		}
	default:
		b := util.I64ToB(int64(n))
		for i := 0; i < 8; i++ {
			(*z)[p+i] = b[i]
		}
	}
}

func (z *ZipListImpl) loadInteger(p, len int) int {
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

func (z *ZipListImpl) storeString(p int, s []byte) {
	for i := 0; i < len(s); i++ {
		(*z)[p+i] = s[i]
	}
}

func (z *ZipListImpl) loadString(p, len int) []byte {
	return (*z)[p : p+len]
}

// zipListInsert inserts the element at the tail.
// This is different from Redis implementation as Redis supports insertion at given index other than the tail.
func (z *ZipListImpl) zipListInsert(e interface{}) error {
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
		return ZLEntryExceedLimitErr
	}

	// resize and update ZLBytes, ZLTail
	newLen = curLen + reqLen
	z.Append(make([]byte, reqLen)...)
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
func (z *ZipListImpl) decodePrevLen(p int) (prevLenSize uint8, prevLen int) {
	if (*z)[p] < ZL_BIG_PREVLEN {
		prevLenSize = 1
		prevLen = int(util.BToUI8(*z, p))
	} else {
		prevLenSize = 5
		prevLen = int(util.BToUI32(*z, p+1))
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

func (z *ZipListImpl) newZipListEntry(p int) (e zipListEntry) {
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
