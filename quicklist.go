package main

type QuickList interface {
	
	AddFirst(interface{}) bool
	AddLast(interface{}) bool
	
	RemoveFirst(interface{}) bool
	RemoveLast(interface{}) bool

	Iterate() interface{}
}

type QuickListImpl struct {
	Head   *QuickListNode // 8
	Tail   *QuickListNode // 8
	Count  int            // 8
	Length uint32         // 4
	Fill   int16          // 4
}

const (
	QUICKLIST_FILL_OPTION = -2
)

var (
	QuickList_ZL_Max_Sizes = [5]int{4096, 8192, 16384, 32768, 65536}
)

type QuickListNode struct {
	Prev *QuickListNode // 8
	Next *QuickListNode // 8
	ZL   *ZipList       // 8
	SZ   uint32         // 4

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
	BitFields uint32 // 4
}
