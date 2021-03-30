package object

const (
	OBJ_STRING = 0
	OBJ_LIST   = 1
	OBJ_SET    = 2
	OBJ_ZSET   = 3
	OBJ_HASH   = 4

	OBJ_ENCODING_STR       = 0 // string
	OBJ_ENCODING_INT       = 1 // int64
	OBJ_ENCODING_HT        = 2 // hash table
	OBJ_ENCODING_ZIPLIST   = 3 // ziplist
	OBJ_ENCODING_INTSET    = 4 // intset
	OBJ_ENCODING_SKIPLIST  = 5 // skiplist
	OBJ_ENCODING_QUICKLIST = 6 // quicklist
)

type ValueObject struct {
	Type      uint8       // objectType + encodingTye
	LRU       uint32      // lru time
	Structure interface{} // internal data structure
}

func (o *ValueObject) SetType(ot, et uint8) {
	o.Type = (ot << 4) | et
}

func (o *ValueObject) GetType() (ot, et uint8) {
	ot = o.Type >> 4
	et = o.Type & 0x0F
	return
}
