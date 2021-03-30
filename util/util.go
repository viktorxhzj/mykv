package util

type Iterator interface {
	Reset()
	Next() interface{}
}

func AssertValidType(e interface{}) (ss []byte, ii, t int) {
	if s, ok := e.(string); ok {
		ss = []byte(s)
	} else if i, ok := e.(int); ok {
		ii = i
		t = 1
	} else {
		t = -1
	}
	return
}