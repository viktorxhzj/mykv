package datastructure

import (
	"errors"
	"math/rand"
	"strings"
)

type SkipList interface {
	Add(string, float64) error
	Delete(string, float64) error
	Contains(string, float64) bool
	GetRank(string, float64) int
	Size() int
}

const (
	SL_MAX_LEVEL   = 32
	SL_PROBABILITY = 0.25
)

var (
	ErrSLInputNotFound = errors.New("no given input has been found")
)

type SkipListImpl struct {
	Head  *SkipListNode
	Tail  *SkipListNode
	Len   int
	Level int
}

type SkipListNode struct {
	Key      string
	Score    float64
	Backward *SkipListNode
	Levels   []SkipListLevel
}

type SkipListLevel struct {
	Span    int
	Forward *SkipListNode
}

func NewSkipListNode(key string, score float64) *SkipListNode {
	node := new(SkipListNode)
	node.Key = key
	node.Score = score
	return node
}

func (node *SkipListNode) initLevels(level int) {
	node.Levels = make([]SkipListLevel, level)
}

func NewSkipList() SkipList {
	sl := new(SkipListImpl)
	sl.Level = 1
	sl.Len = 0
	sl.Head = NewSkipListNode("", 0.0)
	sl.Head.initLevels(SL_MAX_LEVEL)
	return sl
}

func (sl *SkipListImpl) Size() int {
	return sl.Len
}


func (sl *SkipListImpl) Add(key string, score float64) error {
	var rank [SL_MAX_LEVEL]int
	var update [SL_MAX_LEVEL]*SkipListNode
	x := sl.Head

	for i := sl.Level - 1; i >= 0; i-- {
		if i == sl.Level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		if x.Levels[i].Forward != nil && (x.Levels[i].Forward.Score == score && key != x.Levels[i].Forward.Key) {
			return ErrDuplicateInput
		}
		for x.Levels[i].Forward != nil && (x.Levels[i].Forward.Score < score || (x.Levels[i].Forward.Score == score && strings.Compare(key, x.Levels[i].Forward.Key) > 0)) {

			rank[i] += x.Levels[i].Span
			x = x.Levels[i].Forward
		}
		update[i] = x
	}

	curLevel := randomLevel()
	x = NewSkipListNode(key, score)
	x.initLevels(curLevel)

	if curLevel > sl.Level {
		for i := sl.Level; i < curLevel; i++ {
			rank[i] = 0
			update[i] = sl.Head
			update[i].Levels[i].Span = sl.Len
		}
		sl.Level = curLevel
	}

	for i := 0; i < curLevel; i++ {
		x.Levels[i].Forward = update[i].Levels[i].Forward
		update[i].Levels[i].Forward = x

		x.Levels[i].Span = update[i].Levels[i].Span - (rank[0] - rank[i])
		update[i].Levels[i].Span = rank[0] - rank[i] + 1
	}

	for i := sl.Level; i < curLevel; i++ {
		update[i].Levels[i].Span++
	}

	if update[0] != sl.Head {
		x.Backward = update[0]
	}
	if x.Levels[0].Forward != nil {
		x.Levels[0].Forward.Backward = x
	} else {
		sl.Tail = x
	}

	sl.Len++
	return nil
}

func (sl *SkipListImpl) Delete(key string, score float64) error {
	var update [SL_MAX_LEVEL]*SkipListNode
	x := sl.Head

	for i := sl.Level - 1; i >= 0; i-- {
		for x.Levels[i].Forward != nil && (x.Levels[i].Forward.Score < score || (x.Levels[i].Forward.Score == score && strings.Compare(key, x.Levels[i].Forward.Key) > 0)) {
			x = x.Levels[i].Forward
		}
		update[i] = x
	}

	x = x.Levels[0].Forward

	if x != nil && x.Key == key && x.Score == score {
		sl.deleteNode(x, update)
		return nil
	}

	return ErrSLInputNotFound
}

func (sl *SkipListImpl) Contains(key string, score float64) bool {
	x := sl.Head
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Levels[i].Forward != nil && (x.Levels[i].Forward.Score < score || (x.Levels[i].Forward.Score == score && strings.Compare(key, x.Levels[i].Forward.Key) > 0)) {
			x = x.Levels[i].Forward
		}
	}

	x = x.Levels[0].Forward

	if x != nil && x.Score == score && x.Key == key {
		return true
	}
	return false
}

func (sl *SkipListImpl) GetRank(key string, score float64) int {
	var rank int
	x := sl.Head
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Levels[i].Forward != nil && (x.Levels[i].Forward.Score < score || (x.Levels[i].Forward.Score == score && strings.Compare(key, x.Levels[i].Forward.Key) >= 0)) {
			rank += x.Levels[i].Span
			x = x.Levels[i].Forward
		}
		if x.Key != "" && x.Key == key {
			return rank - 1
		}
	}
	return -1
}

func (sl *SkipListImpl) deleteNode(x *SkipListNode, update [SL_MAX_LEVEL]*SkipListNode) {

	for i := 0; i < sl.Level; i++ {
		if update[i].Levels[i].Forward == x {
			update[i].Levels[i].Span += x.Levels[i].Span - 1
			update[i].Levels[i].Forward = x.Levels[i].Forward
		} else {
			update[i].Levels[i].Span--
		}
	}

	if x.Levels[0].Forward != nil {
		x.Levels[0].Forward.Backward = x.Backward
	} else {
		sl.Tail = x.Backward
	}

	for sl.Level > 1 && sl.Head.Levels[sl.Level-1].Forward == nil {
		sl.Head.Levels[sl.Level-1].Span = 0
		sl.Level--
	}

	sl.Len--
}

func randomLevel() int {
	level := 1
	for rand.Float64() < SL_PROBABILITY && level < SL_MAX_LEVEL {
		level++
	}
	return level
}
