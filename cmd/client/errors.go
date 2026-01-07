package main

import (
	"errors"
)

var (
	ErrCommonClientError = errors.New("common client error")
	ErrNoteNotFound      = errors.New("note not found")
	ErrNoteAlreadyExist  = errors.New("note already exist")
)
