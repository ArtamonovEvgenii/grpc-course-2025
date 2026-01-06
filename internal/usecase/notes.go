package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/entity"
)

type repository interface {
	InsertNote(ctx context.Context, note entity.Note) error
	ListNotes(ctx context.Context) ([]entity.NoteListData, error)
	GetNote(ctx context.Context, noteUUID entity.NoteUUID) (entity.Note, error)
	UpdateNote(ctx context.Context, noteUpdate entity.NoteUpdateData) error
	DeleteNote(ctx context.Context, noteUUID entity.NoteUUID) error
}

type Notes struct {
	lgr        *slog.Logger
	repository repository
}

func NewNotes(
	lgr *slog.Logger,
	repository repository,
) *Notes {
	notes := &Notes{
		lgr:        lgr,
		repository: repository,
	}

	return notes
}

func (n *Notes) CreateNote(ctx context.Context, noteInit entity.NoteInitialData) (entity.NoteUUID, error) {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return entity.NoteUUID(uuid.Nil), fmt.Errorf("generate new uuid: %w", err)
	}

	note := entity.Note{
		UUID:      entity.NoteUUID(newUUID),
		Title:     noteInit.Title,
		Text:      noteInit.Text,
		CratedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}

	if err = n.repository.InsertNote(ctx, note); err != nil {
		return entity.NoteUUID(uuid.Nil), fmt.Errorf("insert note: %w", err)
	}

	return note.UUID, nil
}

func (n *Notes) ListNotes(ctx context.Context) ([]entity.NoteListData, error) {
	notes, err := n.repository.ListNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}

	return notes, nil
}

func (n *Notes) GetNote(ctx context.Context, noteUUID entity.NoteUUID) (entity.Note, error) {
	note, err := n.repository.GetNote(ctx, noteUUID)
	if err != nil {
		return entity.Note{}, fmt.Errorf("get note: %w", err)
	}

	return note, nil
}

func (n *Notes) UpdateNote(ctx context.Context, noteUpdate entity.NoteUpdateData) error {
	err := n.repository.UpdateNote(ctx, noteUpdate)
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}

	return nil
}

func (n *Notes) DeleteNote(ctx context.Context, noteUUID entity.NoteUUID) error {
	err := n.repository.DeleteNote(ctx, noteUUID)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	return nil
}
