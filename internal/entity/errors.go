package entity

import (
	"errors"
)

var (
	ErrNoteNotFound     = errors.New("note not found")
	ErrNoteAlreadyExist = errors.New("note already exist")
)
