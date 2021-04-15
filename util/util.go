package util

type Iterator interface {
	Reset()
	Next() interface{}
}

func AssertValidType(e interface{}) (ss string, ii, t int) {
	if s, ok := e.(string); ok {
		ss = s
	} else if i, ok := e.(int); ok {
		ii = i
		t = 1
	} else {
		t = -1
	}
	return
}