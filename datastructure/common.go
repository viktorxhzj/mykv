package datastructure

import "errors"

var (
	ErrExceedLimit    = errors.New("reaches maximum size")
	ErrEmpty          = errors.New("empty")
	ErrInvalidIdx     = errors.New("index is out of range")
	ErrDuplicateInput = errors.New("the input already exists")
)
