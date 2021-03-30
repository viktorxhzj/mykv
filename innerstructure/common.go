package innerstructure

import "errors"

var (
	ExceedLimitErr    = errors.New("reaches maximum size")
	EmptyErr          = errors.New("empty")
	InvalidIdxErr     = errors.New("index is out of range")
	DuplicateInputErr = errors.New("the input already exists")
)
