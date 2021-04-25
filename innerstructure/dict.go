package innerstructure

type Dict struct {
	rehashIdx int64
	table     [2]dictTable
}

const (
	DICT_OK              = 0
	DICT_ERR             = 1
	DICT_TABLE_INIT_SIZE = 4
	DICT_RESIZE_RATIO    = 5
)

type dictTable struct {
	Entries  []*dictEntry
	Size     int64
	SizeMask int64
	Used     int64
}

// Key and Val must be comparable types.
type dictEntry struct {
	Key  string
	Val  interface{}
	Next *dictEntry
}

func NewDict() *Dict {
	d := new(Dict)
	d.rehashIdx = -1
	return d
}

func (d *Dict) isRehashing() bool {
	return d.rehashIdx != -1
}

func (d *Dict) Put(key string, val interface{}) error {
	hash := stringHash(key)
	e := d.addRaw(key, hash)
	e.Key = key
	e.Val = val
	return nil
}

func (d *Dict) Get(key string) interface{} {
	hash := stringHash(key)
	if d.Size() == 0 {
		return nil
	}
	if d.isRehashing() {
		d.rehashStep()
	}
	for i := 0; i < 2; i++ {
		idx := hash & d.table[i].SizeMask
		he := d.table[i].Entries[idx]
		for he != nil {
			if key == he.Key {
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

func (d *Dict) Size() int {
	return int(d.table[0].Used + d.table[1].Used)
}

func (d *Dict) rehashStep() {
	d.rehash(1)
}

// return true if completes
func (d *Dict) rehash(n int) bool {
	emptyVisits := n * 10
	if !d.isRehashing() {
		return true
	}

	for n > 0 && (d.table[0].Used > 0) {
		n--
		for d.table[0].Entries[d.rehashIdx] == nil {
			d.rehashIdx++
			emptyVisits--
			if emptyVisits == 0 {
				return false
			}
		}
		de := d.table[0].Entries[d.rehashIdx]
		for de != nil {
			nxt := de.Next
			h := stringHash(de.Key) & d.table[1].SizeMask
			de.Next = d.table[1].Entries[h]
			d.table[1].Entries[h] = de
			d.table[1].Used++
			d.table[0].Used--
			de = nxt
		}
		d.table[0].Entries[d.rehashIdx] = nil
		d.rehashIdx++
	}

	if d.table[0].Used == 0 {
		d.TransferTable()
		d.rehashIdx = -1
		return true
	}
	return false
}

func (d *Dict) TransferTable() {
	d.table[0].Entries = d.table[1].Entries
	d.table[0].Size = d.table[1].Size
	d.table[0].SizeMask = d.table[1].SizeMask
	d.table[0].Used = d.table[1].Used
	d.table[1].Entries = nil
	d.table[1].Size = 0
	d.table[1].SizeMask = 0
	d.table[1].Used = 0
}

func (d *Dict) addRaw(key interface{}, hash int64) *dictEntry {

	if d.isRehashing() {
		d.rehashStep()
	}

	idx, e := d.keyIndex(key, hash)
	if e != nil {
		return e
	}

	ne := new(dictEntry)
	var i int

	if d.isRehashing() {
		i = 1
	} else {
		i = 0
	}

	ne.Next = d.table[i].Entries[idx]
	d.table[i].Entries[idx] = ne
	d.table[i].Used++
	return ne
}

func (d *Dict) keyIndex(key interface{}, hash int64) (int64, *dictEntry) {

	if d.expandIfNeeded() == DICT_ERR {
		return -1, nil
	}
	var idx int64
	for i := 0; i < 2; i++ {
		idx = hash & d.table[i].SizeMask
		he := d.table[i].Entries[idx]
		for he != nil {
			if key == he.Key {
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

func (d *Dict) expandIfNeeded() int {

	// Incremental rehashing already in progress. Return.
	if d.isRehashing() {
		return DICT_OK
	}

	if d.table[0].Size == 0 {
		return d.expand(DICT_TABLE_INIT_SIZE)
	}

	if d.table[0].Used >= d.table[0].Size {
		return d.expand(d.table[0].Used + 1)
	}

	return DICT_OK

}

func (d *Dict) expand(size int64) int {

	// expand shouldn't take place when rehashing
	// invalid size
	if d.isRehashing() || d.table[0].Used > size {
		return DICT_ERR
	}

	realSize := nextPower(size)

	if realSize == d.table[0].Size {
		return DICT_ERR
	}

	entries := make([]*dictEntry, realSize)

	var i int
	if d.table[0].Entries != nil {
		i = 1
		d.rehashIdx = 0
	}

	d.table[i].Entries = entries
	d.table[i].Size = realSize
	d.table[i].SizeMask = realSize - 1

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