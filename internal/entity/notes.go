package entity

import (
	"time"

	"github.com/google/uuid"
)

type NoteUUID uuid.UUID

func (n *NoteUUID) String() string {
	return uuid.UUID(*n).String()
}

type Note struct {
	UUID      NoteUUID
	Title     string
	Text      string
	CratedAt  time.Time
	UpdatedAt time.Time
}

type NoteInitialData struct {
	Title string
	Text  string
}

type NoteListData struct {
	UUID      NoteUUID
	Title     string
	UpdatedAt time.Time
}

type NoteUpdateData struct {
	UUID  NoteUUID
	Title string
	Text  string
}
