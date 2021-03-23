package main

import (
	"fmt"
)

const (
	ZIPLIST_MAX_LENGTH = 0xFFFF

	ZIPLIST_ZLTAIL_OFFSET = 4
	ZIPLIST_ZLLEN_OFFSET  = 8
	ZIPLIST_INIT_SIZE     = 11
	ZIPLIST_ZLEND         = 0xFF

	ZIPLISTENTRY_1BYTE_MAX_SIZE = 1<<6 - 1

	ZIPLISTENTRY_2BYTES_MAX_SIZE = 1<<14 - 1

	// a ZipListEntry can use uint32 to represent the length of the previous entry.
	// the PrevLen part takes 5 bytes at most, the Encoding part takes 5 bytes at most.
	// as a result, the entry part can take uint32 - 10 bytes at most.
	ZIPLISTENTRY_5BYTES_MAX_SIZE = 1<<32 - 11
)

// <zlbytes> <zltail> <zllen> <entry> <entry> ... <entry> <zlend>
// uint32    uint32   uint16							  uint8
type ZipList struct {
	List []byte
}

func NewZipList() *ZipList {
	z := new(ZipList)
	z.List = append(z.List, UI32ToB(ZIPLIST_INIT_SIZE)...)
	z.List = append(z.List, UI32ToB(0)...)
	z.List = append(z.List, UI16ToB(0)...)
	z.List = append(z.List, UI8ToB(ZIPLIST_ZLEND)...)

	return z
}

func (z *ZipList) ZLBytes() uint32 {
	return uint32(len(z.List))
}

func (z *ZipList) ZLTail() uint32 {
	return BToUI32(z.List, ZIPLIST_ZLTAIL_OFFSET)
}

func (z *ZipList) ZLLen() uint16 {
	return BToUI16(z.List, ZIPLIST_ZLLEN_OFFSET)
}

func (z *ZipList) Push(data interface{}) bool {
	s, ok1 := data.(string)
	i, ok2 := data.(int)

	if z.ZLLen() == ZIPLIST_MAX_LENGTH {
		fmt.Println("zip list reaches the maximum size")
		return false
	}

	if !ok1 && !ok2 {
		fmt.Println("input data is either string or integer")
		return false
	}

	var prevLen, encoding, entry []byte

	if ok1 {
		if len([]byte(s)) >= ZIPLISTENTRY_5BYTES_MAX_SIZE {
			fmt.Println("input string is too long")
			return false
		}
		encoding, entry = z.encodeString(s)
	}

	if ok2 {
		encoding, entry = z.encodeInteger(i)
	}

	prevLen = z.calculatePrevLen()

	z.updateZLTail()

	z.List = z.List[:len(z.List)-1]
	z.List = append(z.List, prevLen...)
	z.List = append(z.List, encoding...)
	z.List = append(z.List, entry...)
	z.List = append(z.List, UI8ToB(ZIPLIST_ZLEND)...)

	z.updateZLLen()

	return true
}

func (z *ZipList) All() (res []interface{}) {
	if l := z.ZLLen(); l == 0 {
		return
	} else {
		res = make([]interface{}, l)

		c := z.ZLTail()

		var prevLen, currLen uint32
		currLen = z.ZLBytes() - c - 1

		for i := 0; i < int(l); i++ {
			// PrevLen consumes either 1 byte or 5 bytes.
			var offset uint32

			if z.List[c] != 0xFE {
				prevLen = uint32(z.List[c])
				offset += 1
			} else {
				prevLen = BToUI32(z.List, int(c+1))
				offset += 5
			}

			if z.List[c+offset] >> 6 <= 2 {
				res[int(l)-i-1] = z.decodeString(c, offset, currLen)

			} else {
				res[int(l)-i-1] = z.decodeInteger(c, offset, currLen)
			}
			c -= prevLen
			currLen = prevLen

		}
	}
	return
}

func (z *ZipList) Index(data interface{}) int {
	ss, ok1 := data.(string)
	ii, ok2 := data.(int)

	if !ok1 && !ok2 {
		fmt.Println("input data is either string or integer")
		return -1
	}

	if l := z.ZLLen(); l == 0 {
		return -1
	} else {
		c := z.ZLTail()

		var prevLen, currLen uint32
		currLen = z.ZLBytes() - c - 1

		for i := 0; i < int(l); i++ {
			var offset uint32

			if z.List[c] != 0xFE {
				prevLen = uint32(z.List[c])
				offset += 1
			} else {
				prevLen = BToUI32(z.List, int(c+1))
				offset += 5
			}

			if z.List[c+offset] >> 6 <= 2 {
				if ss == z.decodeString(c, offset, currLen) {
					return int(l)-i-1
				}

			} else {
				if ii == z.decodeInteger(c, offset, currLen) {
					return int(l)-i-1
				}
			}
			c -= prevLen
			currLen = prevLen
		}
	}
	return -1
}

func (z *ZipList) Insert(data interface{}) {

}

func (z *ZipList) updateZLTail() {
	b := UI32ToB(z.ZLBytes() - 1)
	for i := 0; i < 4; i++ {
		z.List[ZIPLIST_ZLTAIL_OFFSET+i] = b[i]
	}
}

func (z *ZipList) updateZLLen() {
	b := UI16ToB(z.ZLLen() + 1)
	for i := 0; i < 2; i++ {
		z.List[ZIPLIST_ZLLEN_OFFSET+i] = b[i]
	}
}

func (z *ZipList) encodeString(s string) (encoding, entry []byte) {
	entry = []byte(s)
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

func (z *ZipList) decodeString(c, offset, currLen uint32) string {
	switch {
	case z.List[c+offset] >> 6 == 0x00:
		offset += 1
	case z.List[c+offset] >> 6 == 0x01:
		offset += 2
	case z.List[c+offset] >> 6 == 0x02:
		offset += 5
	}
		
	return string(z.List[c+offset : c+currLen])
} 

func (z *ZipList) encodeInteger(i int) (encoding, entry []byte) {
	switch {
	case i >= 0 && i <= 13:
		encoding = UI8ToB(0xF0 | uint8(i))
	
	case i >= -1<<7 && i <= 1<<7-1:
		encoding = UI8ToB(0xFE)
		entry = I8ToB(int8(i))
	
	case i >= -1<<15 && i <= 1<<15-1:
		encoding = UI8ToB(0xC0)
		entry = I16ToB(int16(i))
	
	case i >= -1<<31 && i <= 1<<31-1:
		encoding = UI8ToB(0xD0)
		entry = I32ToB(int32(i))
	
	case i >= -1<<63 && i <= 1<<63-1:
		encoding = UI8ToB(0xE0)
		entry = I64ToB(int64(i))
	}
	return
}

func (z *ZipList) decodeInteger(c, offset, currLen uint32) interface{} {
	var i int
	switch z.List[c+offset] {
	case 0xFE:
		offset += 1
		i = int(int8(z.List[c+offset]))

	case 0xC0:
		offset += 1
		i = int(BToI16(z.List, int(c+offset)))

	case 0xD0:
		offset += 1
		i = int(BToI32(z.List, int(c+offset)))

	case 0xE0:
		offset += 1
		i = int(BToI64(z.List, int(c+offset)))

	default:
		i = int(uint8(z.List[c+offset] & 0x0F))
	}
	return i
}

func (z *ZipList) calculatePrevLen() (prevLen []byte) {
	
	tailOffset := z.ZLTail()

	if tailOffset == 0 {
		prevLen = make([]byte, 1)
	} else {
		p := z.ZLBytes()-1-tailOffset
		if p < 0xFE {
			prevLen = UI8ToB(uint8(p))
		} else {
			prevLen = UI8ToB(0xFE)
			prevLen = append(prevLen, UI32ToB(z.ZLBytes()-1-tailOffset)...)
		}
	}
	return
}