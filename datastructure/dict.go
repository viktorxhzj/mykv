package datastructure

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

type dictEntry struct {
	Key  string
	Val  string
	Next *dictEntry
}

func NewDict() *Dict {
	d := new(Dict)
	d.rehashIdx = -1
	return d
}

// Put puts a key-pair into the dictionary.
func (d *Dict) Put(key, val string) {
	hash := stringHash(key)
	e := d.addRaw(key, hash)
	e.Key = key
	e.Val = val
}

// Get gets the value designated by the key.
// If the key-pair is not in the dictionary, it returns "".
func (d *Dict) Get(key string) string {
	if d.Size() == 0 {
		return ""
	}

	hash := stringHash(key)

	// if is rehashing, we rehash first.
	if d.isRehashing() {
		d.rehash(1)
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
		// if is rehashing, both tables are functioning.
		// we need to check the second table as well.
		if !d.isRehashing() {
			return ""
		}
	}
	return ""
}

func (d *Dict) Delete(key string) {
	if d.Size() == 0 {
		return
	}
	if d.isRehashing() {
		d.rehash(1)
	}
	hash := stringHash(key)
	for i := 0; i < 2; i++ {
		idx := hash & d.table[i].SizeMask
		he := d.table[i].Entries[idx]
		if he != nil && key == he.Key {
			d.table[i].Entries[idx] = he.Next
			d.table[i].Used--
			return
		}
		for he != nil && he.Next != nil {
			if key == he.Next.Key {
				he.Next = he.Next.Next
				d.table[i].Used--
				return
			}
			he = he.Next
		}

		// if is rehashing, we need to search for the second table.
		if !d.isRehashing() {
			break
		}
	}
}

func (d *Dict) Size() int {
	return int(d.table[0].Used + d.table[1].Used)
}

func (d *Dict) isRehashing() bool {
	return d.rehashIdx != -1
}

// rehash rehashes atmost n buckets.
// it returns true if rehash is completed.
func (d *Dict) rehash(n int) bool {
	emptyVisits := n * 10

	for n > 0 && (d.table[0].Used > 0) {
		n--
		for d.table[0].Entries[d.rehashIdx] == nil {
			d.rehashIdx++
			emptyVisits--
			if emptyVisits == 0 {
				return false
			}
		}
		// transfer a bucket from the old table to the new table
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
		d.transferTable()
		d.rehashIdx = -1
		return true
	}
	return false
}

func (d *Dict) transferTable() {
	d.table[0].Entries = d.table[1].Entries
	d.table[0].Size = d.table[1].Size
	d.table[0].SizeMask = d.table[1].SizeMask
	d.table[0].Used = d.table[1].Used
	d.table[1].Entries = nil
	d.table[1].Size = 0
	d.table[1].SizeMask = 0
	d.table[1].Used = 0
}

// addRaw returns the pointer to the entry of the given index.
func (d *Dict) addRaw(key string, hash int64) *dictEntry {
	if d.isRehashing() {
		d.rehash(1)
	}
	idx, e := d.keyIndex(key, hash)
	if e != nil || idx == -1 {
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

// keyIndex returns the index of the key at the hash table,
// and the pointer to the entry. if there is no such a key, the pointer is nil.
func (d *Dict) keyIndex(key string, hash int64) (int64, *dictEntry) {

	d.expandIfNeeded()
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

		// if is rehashing, we need to search for the second table.
		if !d.isRehashing() {
			break
		}
	}
	return idx, nil
}

func (d *Dict) expandIfNeeded() {

	// Incremental rehashing already in progress. Return.
	if d.isRehashing() {
		return
	}

	if d.table[0].Size == 0 {
	// upon initialization.
		d.expand(DICT_TABLE_INIT_SIZE)
	} else if d.table[0].Used >= d.table[0].Size {
	// loading factor > 1.
		d.expand(d.table[0].Used + 1)
	}
}

// expand expands the dictionary
func (d *Dict) expand(size int64) {

	realSize := nextPower(size)
	entries := make([]*dictEntry, realSize)

	var i int
	if d.table[0].Entries != nil {
		i = 1
		d.rehashIdx = 0
	}

	d.table[i].Entries = entries
	d.table[i].Size = realSize
	d.table[i].SizeMask = realSize - 1
}

// nextPower returns a 2-power size starting from 4.
func nextPower(size int64) int64 {
	i := int64(DICT_TABLE_INIT_SIZE)
	for {
		if i >= size {
			return i
		}
		i *= 2
	}
}

// stringHash returns the 64-bit hash of the string.
func stringHash(str string) int64 {
	var h int64

	b := []byte(str)

	for i := 0; i < len(b); i++ {
		h = 31*h + int64(b[i])
	}
	return h
}
