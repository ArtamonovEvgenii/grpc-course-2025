package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/entity"
)

type Storage struct {
	data  map[entity.NoteUUID]entity.Note
	mutex sync.RWMutex
}

func NewStorage() *Storage {
	storage := &Storage{
		data:  make(map[entity.NoteUUID]entity.Note),
		mutex: sync.RWMutex{},
	}

	return storage
}

func (s *Storage) InsertNote(ctx context.Context, note entity.Note) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exist := s.data[note.UUID]; exist {
		return entity.ErrNoteAlreadyExist
	}

	s.data[note.UUID] = note

	return nil
}

func (s *Storage) ListNotes(_ context.Context) ([]entity.NoteListData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	res := make([]entity.NoteListData, 0, len(s.data))
	for _, note := range s.data {
		noteListElement := entity.NoteListData{
			UUID:      note.UUID,
			Title:     note.Title,
			UpdatedAt: note.UpdatedAt,
		}

		res = append(res, noteListElement)
	}

	return res, nil
}

func (s *Storage) GetNote(_ context.Context, noteUUID entity.NoteUUID) (entity.Note, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	note, exists := s.data[noteUUID]
	if !exists {
		return entity.Note{}, fmt.Errorf("%w: uuid %s", entity.ErrNoteNotFound, noteUUID.String())
	}

	return note, nil
}

func (s *Storage) UpdateNote(_ context.Context, noteUpdate entity.NoteUpdateData) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	prevNote, exist := s.data[noteUpdate.UUID]
	if !exist {
		return fmt.Errorf("%w: uuid %s", entity.ErrNoteNotFound, noteUpdate.UUID.String())
	}

	newNote := entity.Note{
		UUID:      prevNote.UUID,
		Title:     noteUpdate.Title,
		Text:      noteUpdate.Text,
		CratedAt:  prevNote.CratedAt,
		UpdatedAt: time.Now(),
	}

	s.data[prevNote.UUID] = newNote

	return nil
}

func (s *Storage) DeleteNote(_ context.Context, noteUUID entity.NoteUUID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exist := s.data[noteUUID]
	if !exist {
		return fmt.Errorf("%w: uuid %s", entity.ErrNoteNotFound, noteUUID.String())
	}

	delete(s.data, noteUUID)

	return nil
}
