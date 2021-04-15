package util

//
// bytes to uint
//

func BToUI64(b []byte, offset int) (res uint64) {
	for i := 0; i < 8; i++ {
		res |= uint64(b[offset]) << (8 * (7 - i))
		offset++
	}
	return
}

func BToUI32(b []byte, offset int) (res uint32) {
	for i := 0; i < 4; i++ {
		res |= uint32(b[offset]) << (8 * (3 - i))
		offset++
	}
	return
}

func BToUI16(b []byte, offset int) (res uint16) {
	for i := 0; i < 2; i++ {
		res |= uint16(b[offset]) << (8 * (1 - i))
		offset++
	}
	return
}

func BToUI8(b []byte, offset int) (res uint8) {
	res = uint8(b[offset])
	return
}

//
// bytes to int
//

func BToI64(b []byte, offset int) (res int64) {
	for i := 0; i < 8; i++ {
		res |= int64(b[offset]) << (8 * (7 - i))
		offset++
	}
	return
}

func BToI32(b []byte, offset int) (res int32) {
	for i := 0; i < 4; i++ {
		res |= int32(b[offset]) << (8 * (3 - i))
		offset++
	}
	return
}

func BToI16(b []byte, offset int) (res int16) {
	for i := 0; i < 2; i++ {
		res |= int16(b[offset]) << (8 * (1 - i))
		offset++
	}
	return
}

func BToI8(b []byte, offset int) (res int8) {
	res = int8(b[offset])
	return
}

//
// uint to bytes
//

func UI64ToB(n uint64, res []byte, offset int) {
	for i := 0; i < 8; i++ {
		res[i+offset] = byte((n >> (8 * (7 - i))) & 0xFF)
	}
}

func UI32ToB(n uint32, res []byte, offset int) {
	for i := 0; i < 4; i++ {
		res[i+offset] = byte((n >> (8 * (3 - i))) & 0xFF)
	}
}

func UI16ToB(n uint16, res []byte, offset int) {
	for i := 0; i < 2; i++ {
		res[i+offset] = byte((n >> (8 * (1 - i))) & 0xFF)
	}
}

func UI8ToB(n uint8, res []byte, offset int) {
	res[offset] = n
}

//
// int to bytes
//

func I64ToB(n int64, res []byte, offset int) {
	for i := 0; i < 8; i++ {
		res[i+offset] = byte((n >> (8 * (7 - i))) & 0xFF)
	}
}

func I32ToB(n int32, res []byte, offset int) {
	for i := 0; i < 4; i++ {
		res[i+offset] = byte((n >> (8 * (3 - i))) & 0xFF)
	}
}

func I16ToB(n int16, res []byte, offset int) {
	for i := 0; i < 2; i++ {
		res[i+offset] = byte((n >> (8 * (1 - i))) & 0xFF)
	}
}

func I8ToB(n int8, res []byte, offset int) {
	res[offset] = byte(n)
}
