package innerstructure


import (
	"errors"
	"math"
)

type Dict interface {
	Put(interface{}, interface{}) error
	Get(interface{}) interface{}
	Size() int
}

const (
	DICT_OK              = 0
	DICT_ERR             = 1
	DICT_TABLE_INIT_SIZE = 4
	DICT_RESIZE_RATIO    = 5
)

var (
	DictInvalidKeyErr = errors.New("input key is invalid")
)

type DictKey interface {
	Compare(DictKey) bool
	Hash() int64
}

type DictImpl struct {
	RehashIdx int64
	Table     [2]DictTable
}

type DictTable struct {
	Entries  []*DictEntry
	Size     int64
	SizeMask int64
	Used     int64
}

// Key and Val must be comparable types.
type DictEntry struct {
	Key  interface{}
	Val  interface{}
	Next *DictEntry
}

func NewDict() Dict {
	d := new(DictImpl)
	d.RehashIdx = -1
	return d
}

func (d *DictImpl) isRehashing() bool {
	return d.RehashIdx != -1
}

func (d *DictImpl) Put(key interface{}, val interface{}) error {

	hash := hashKey(key)
	if hash == -1 {
		return DictInvalidKeyErr
	}

	e := d.addRaw(key, hash)
	e.Key = key
	e.Val = val
	return nil
}

func (d *DictImpl) Get(key interface{}) interface{} {
	hash := hashKey(key)
	if hash == -1 {
		return DictInvalidKeyErr
	}
	if d.Size() == 0 {
		return nil
	}
	if d.isRehashing() {
		d.rehashStep()
	}
	for i := 0; i < 2; i++ {
		idx := hash & d.Table[i].SizeMask
		he := d.Table[i].Entries[idx]
		for he != nil {
			if compareKeys(key, he.Key) {
				return he.Val
			}
			he = he.Next
		}

		if !d.isRehashing() {
			return nil
		}
	}

	return nil
}

func (d *DictImpl) Size() int {
	return int(d.Table[0].Used + d.Table[1].Used)
}

func (d *DictImpl) rehashStep() {
	d.rehash(1)
}

// return true if completes
func (d *DictImpl) rehash(n int) bool {
	emptyVisits := n * 10
	if !d.isRehashing() {
		return true
	}

	for n > 0 && (d.Table[0].Used > 0) {
		n--
		for d.Table[0].Entries[d.RehashIdx] == nil {
			d.RehashIdx++
			emptyVisits--
			if emptyVisits == 0 {
				return false
			}
		}
		de := d.Table[0].Entries[d.RehashIdx]
		for de != nil {
			nxt := de.Next
			h := hashKey(de.Key) & d.Table[1].SizeMask
			de.Next = d.Table[1].Entries[h]
			d.Table[1].Entries[h] = de
			d.Table[1].Used++
			d.Table[0].Used--
			de = nxt
		}
		d.Table[0].Entries[d.RehashIdx] = nil
		d.RehashIdx++
	}

	if d.Table[0].Used == 0 {
		d.TransferTable()
		d.RehashIdx = -1
		return true
	}
	return false
}

func (d *DictImpl) TransferTable() {
	d.Table[0].Entries = d.Table[1].Entries
	d.Table[0].Size = d.Table[1].Size
	d.Table[0].SizeMask = d.Table[1].SizeMask
	d.Table[1].Entries = nil
	d.Table[1].Size = 0
	d.Table[1].SizeMask = 0
}

func (d *DictImpl) addRaw(key interface{}, hash int64) *DictEntry {

	if d.isRehashing() {
		d.rehashStep()
	}

	idx, e := d.keyIndex(key, hash)
	if e != nil {
		return e
	}

	ne := new(DictEntry)
	var i int

	if d.isRehashing() {
		i = 1
	} else {
		i = 0
	}

	ne.Next = d.Table[i].Entries[idx]
	d.Table[i].Entries[idx] = ne
	d.Table[i].Used++
	return ne
}

func compareKeys(k1, k2 interface{}) bool {
	i1, ok1 := k1.(int)
	i2, ok2 := k2.(int)

	if ok1 && ok2 {
		return i1 == i2
	}

	s1, ok1 := k1.(string)
	s2, ok2 := k2.(string)

	if ok1 && ok2 {
		return s1 == s2
	}

	c1, ok1 := k1.(DictKey)
	c2, ok2 := k2.(DictKey)

	if ok1 && ok2 {
		return c1.Compare(c2)
	}
	return false
}

func (d *DictImpl) keyIndex(key interface{}, hash int64) (int64, *DictEntry) {

	if d.expandIfNeeded() == DICT_ERR {
		return -1, nil
	}
	var idx int64
	for i := 0; i < 2; i++ {
		idx = hash & d.Table[i].SizeMask
		he := d.Table[i].Entries[idx]
		for he != nil {
			if compareKeys(key, he.Key) {
				return -1, he
			}
			he = he.Next
		}

		if !d.isRehashing() {
			break
		}
	}
	return idx, nil
}

func (d *DictImpl) expandIfNeeded() int {

	// Incremental rehashing already in progress. Return.
	if d.isRehashing() {
		return DICT_OK
	}

	if d.Table[0].Size == 0 {
		return d.expand(DICT_TABLE_INIT_SIZE)
	}

	if d.Table[0].Used >= d.Table[0].Size {
		return d.expand(d.Table[0].Used + 1)
	}

	return DICT_OK

}

func (d *DictImpl) expand(size int64) int {

	// expand shouldn't take place when rehashing
	// invalid size
	if d.isRehashing() || d.Table[0].Used > size {
		return DICT_ERR
	}

	realSize := nextPower(size)

	if realSize == d.Table[0].Size {
		return DICT_ERR
	}

	entries := make([]*DictEntry, realSize)

	var i int
	if d.Table[0].Entries != nil {
		i = 1
		d.RehashIdx = 0
	}

	d.Table[i].Entries = entries
	d.Table[i].Size = realSize
	d.Table[i].SizeMask = realSize - 1

	return DICT_OK
}

func nextPower(size int64) int64 {
	i := int64(DICT_TABLE_INIT_SIZE)

	for {
		if i >= size {
			return i
		}
		i *= 2
	}
}

func stringHash(str string) int64 {
	var h int64

	b := []byte(str)

	for i := 0; i < len(b); i++ {
		h = 31*h + int64(b[i])
	}
	return h
}

func intHash(ii int) int64 {
	if ii > 0 {
		return int64(ii)
	} else if ii == math.MinInt64 {
		return 0
	}
	return int64(-ii)
}

func hashKey(key interface{}) int64 {
	if s, ok := key.(string); ok {
		return stringHash(s)
	} else if i, ok := key.(int); ok {
		return intHash(i)
	} else if k, ok := key.(DictKey); ok {
		return k.Hash()
	}
	return -1
}
