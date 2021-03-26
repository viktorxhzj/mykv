package main

import (
	"math"
)

type QuickList interface {
	PushHead(interface{}) (bool, error)
	PushTail(interface{}) (bool, error)

	Size() int

	Get(int) (QuickListEntry, error)

	// DeleteHead(interface{}) error
	// DeleteTail(interface{}) error
}

// Redis implementation doesn't have dummy head or dummy tail,
// but here we use them.

type QuickListImpl struct {
	Head  *QuickListNode // 8
	Tail  *QuickListNode // 8
	Count int            // 8
	Len   uint32         // 4
	Fill  int16          // 2
}

const (
	QL_FILL_OPTION   = 3
	QL_MAX_LEN       = math.MaxUint32 - 2 // 2 for dummy head/tail
	QL_MAX_SIZE      = math.MaxInt64
	QL_ZL_SIZE_LIMIT = 1 << 13
)

var (
	QLOptimizationLevel = [5]int{4096, 8192, 16384, 32768, 65536}
)

type QuickListNode struct {
	Prev   *QuickListNode // 8
	Next   *QuickListNode // 8
	ZL     ZipList        // 8
	ZLSize uint32         // 4 ziplist size in bytes

	// |Extra 			   |Count			   |
	// |0000 0000 0000 0000|0000 0000 0000 0000|
	// If Fill is specified with a positive value,
	// a quicklist node (ziplist) has at most 32767 entries;
	//
	// If Fill is specified with a negative value,
	// a quicklist node (ziplist) takes at most 65536 bytes,
	//
	// considering a smallest entry of 2 bytes, the node has at most ~ 32k entries.
	// As a result, a bitfield of 16 is enough for Count.
	//
	// Redis implementation has other bit-fields mainly for compression,
	// here we just ignore them (because we don't support compression)
	// and only keep "Count" left.
	Count int16 // 2
}

type QuickListEntry struct {
	List     QuickList
	Node     *QuickListNode
	String   string
	Integer  int
	IsString bool
	Offset   [2]int
}

func NewQuickList() QuickList {
	q := new(QuickListImpl)
	q.Fill = QL_FILL_OPTION
	q.Head, q.Tail = NewQuickListNode(nil), NewQuickListNode(nil)
	q.Head.Next = q.Tail
	q.Tail.Prev = q.Head
	return q
}

func (q *QuickListImpl) Size() int {
	return int(q.Count)
}

func (q *QuickListImpl) Get(idx int) (entry QuickListEntry, err error) {
	if idx >= q.Count || (idx == math.MinInt64 && q.Count != QL_MAX_SIZE) || (-idx > q.Count) {
		err = InvalidIdxErr
		return
	}

	node := q.Head.Next
	var offset int

	for (idx - int(node.Count)) >= 0 {
		idx -= int(node.Count)
		node = node.Next
		offset++
	}

	entry.List = q
	entry.Node = node
	entry.Offset[0] = offset
	entry.Offset[1] = idx

	e, _ := node.ZL.Get(idx)
	if s, ok := e.(string); ok {
		entry.IsString = true
		entry.String = s
	} else if i, ok := e.(int); ok {
		entry.Integer = i
	}
	return
}

func (q *QuickListImpl) PushHead(e interface{}) (created bool, err error) {
	if q.Count == QL_MAX_SIZE {
		err = ExceedLimitErr
		return
	}

	if _, _, t := AssertValidType(e); t == -1 {
		err = ZLInvalidInputErr
		return
	}

	var h *QuickListNode

	if q.Count == 0 {
		h = q.InsertHeadNode()
		created = true
	} else {
		h = q.Head.Next
		if !h.AllowInsert(e, q.Fill) {
			if q.Len == QL_MAX_LEN {
				err = ExceedLimitErr
				return
			}
			h = q.InsertHeadNode()
			created = true
		}
	}

	h.ZL.Add(e)
	h.UpdateSize()
	q.Count++
	h.Count++

	return
}

func (q *QuickListImpl) PushTail(e interface{}) (created bool, err error) {
	if q.Count == QL_MAX_SIZE {
		err = ExceedLimitErr
		return
	}

	if _, _, t := AssertValidType(e); t == -1 {
		err = ZLInvalidInputErr
		return
	}

	var h *QuickListNode

	if q.Count == 0 {
		h = q.InsertTailNode()
		created = true
	} else {
		h = q.Tail.Prev
		if !h.AllowInsert(e, q.Fill) {
			if q.Len == QL_MAX_LEN {
				err = ExceedLimitErr
				return
			}
			h = q.InsertTailNode()
			created = true
		}
	}

	h.ZL.Add(e)
	h.UpdateSize()
	q.Count++
	h.Count++

	return
}

func (q *QuickListImpl) InsertHeadNode() *QuickListNode {
	node := NewQuickListNode(NewZipList())

	nxt := q.Head.Next

	node.Prev = q.Head
	node.Next = nxt

	nxt.Prev = node
	q.Head.Next = node

	q.Len++
	return node
}

func (q *QuickListImpl) InsertTailNode() *QuickListNode {
	node := NewQuickListNode(NewZipList())

	pre := q.Tail.Prev

	node.Prev = pre
	node.Next = q.Tail

	pre.Next = node
	q.Tail.Prev = node

	q.Len++
	return node
}

func NewQuickListNode(z ZipList) *QuickListNode {
	node := new(QuickListNode)
	if z != nil {
		node.ZL = z
	}
	return node
}

func (node *QuickListNode) UpdateSize() {
	node.ZLSize = uint32(node.ZL.ZLBytes())
}

func (node *QuickListNode) AllowInsert(e interface{}, fill int16) bool {
	// estimate offset
	var size, overhead int
	ss, ii, t := AssertValidType(e)

	switch t {
	case 0:
		size = len(ss)
		if size < 254 {
			overhead = 1
		} else {
			overhead = 5
		}

		if size < 1<<6 {
			overhead += 1
		} else if size < 1<<14 {
			overhead += 2
		} else {
			overhead += 5
		}

	case 1:
		overhead = 2
		switch {
		case ii >= 0 && ii <= ZL_INT_IMM_MAX-ZL_INT_IMM_MIN:

		case ii >= math.MinInt8 && ii <= math.MaxInt8:
			overhead += 1
		case ii >= math.MinInt16 && ii <= math.MaxInt16:
			overhead += 2
		case ii >= math.MinInt32 && ii <= math.MaxInt32:
			overhead += 4
		default:
			overhead += 8
		}
	}
	newSize := int(node.ZLSize) + size + overhead

	if quickListNodeMeetsOptimizationRequirement(newSize, fill) {
		return true
	} else if newSize > QL_ZL_SIZE_LIMIT {
		// if fill is positive, then it is first constrained by byte sizes, then by count
		return false
	} else if node.Count < fill {
		return true
	}

	return false
}

func quickListNodeMeetsOptimizationRequirement(size int, fill int16) bool {
	// fill not in range -1 to -5
	if fill >= 0 {
		return false
	}
	offset := int((-fill) - 1)

	// fill in range -1 to -5
	if offset < len(QLOptimizationLevel) {
		if size <= QLOptimizationLevel[offset] {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}
