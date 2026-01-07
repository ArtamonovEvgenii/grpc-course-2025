package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"buf.build/go/protovalidate"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/entity"
	pb "github.com/ArtamonovEvgenii/grpc-course-2025/pkg/api/notes/v1"
)

type notesUsecase interface {
	CreateNote(ctx context.Context, noteInit entity.NoteInitialData) (entity.NoteUUID, error)
	ListNotes(ctx context.Context) ([]entity.NoteListData, error)
	GetNote(ctx context.Context, noteUUID entity.NoteUUID) (entity.Note, error)
	UpdateNote(ctx context.Context, noteUpdate entity.NoteUpdateData) error
	DeleteNote(ctx context.Context, noteUUID entity.NoteUUID) error
}

type Controller struct {
	pb.UnimplementedNotesAPIServer
	lgr          *slog.Logger
	notesUsecase notesUsecase
}

func NewController(
	lgr *slog.Logger,
	notesUsecase notesUsecase,
) (*Controller, error) {
	controller := &Controller{
		lgr:          lgr,
		notesUsecase: notesUsecase,
	}

	return controller, nil
}

func (c *Controller) CreateNote(
	ctx context.Context,
	req *pb.CreateNoteRequest,
) (*pb.CreateNoteResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	noteInit := entity.NoteInitialData{
		Title: req.Title,
		Text:  req.Text,
	}

	noteUUID, err := c.notesUsecase.CreateNote(ctx, noteInit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := &pb.CreateNoteResponse{
		Uuid: noteUUID.String(),
	}

	return resp, nil
}

func (c *Controller) GetNotesList(
	ctx context.Context,
	_ *pb.GetNotesListRequest,
) (*pb.GetNotesListResponse, error) {

	entityNotes, err := c.notesUsecase.ListNotes(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	respNotes := make([]*pb.GetNotesListResponseNote, 0, len(entityNotes))
	for _, entityNote := range entityNotes {
		respNote := &pb.GetNotesListResponseNote{
			Uuid:      entityNote.UUID.String(),
			Title:     entityNote.Title,
			UpdatedAt: responseDateTime(entityNote.UpdatedAt),
		}

		respNotes = append(respNotes, respNote)
	}

	resp := &pb.GetNotesListResponse{
		Notes: respNotes,
	}

	return resp, nil
}

func (c *Controller) GetNote(
	ctx context.Context,
	req *pb.GetNoteRequest,
) (*pb.GetNoteResponse, error) {
	noteUUID, err := uuid.Parse(req.Uuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	note, err := c.notesUsecase.GetNote(ctx, entity.NoteUUID(noteUUID))
	if err != nil {
		return nil, domainToTransportError(err)
	}

	resp := &pb.GetNoteResponse{
		Uuid:      note.UUID.String(),
		Title:     note.Title,
		Text:      note.Text,
		CratedAt:  responseDateTime(note.CratedAt),
		UpdatedAt: responseDateTime(note.UpdatedAt),
	}

	return resp, nil
}

func (c *Controller) UpdateNote(
	ctx context.Context,
	req *pb.UpdateNoteRequest,
) (*pb.UpdateNoteResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	noteUUID, err := uuid.Parse(req.Uuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	noteUpdateData := entity.NoteUpdateData{
		UUID:  entity.NoteUUID(noteUUID),
		Title: req.Title,
		Text:  req.Text,
	}

	err = c.notesUsecase.UpdateNote(ctx, noteUpdateData)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := &pb.UpdateNoteResponse{}

	return resp, nil
}

func (c *Controller) DeleteNote(ctx context.Context, req *pb.DeleteNoteRequest) (*pb.DeleteNoteResponse, error) {
	noteUUID, err := uuid.Parse(req.Uuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = c.notesUsecase.DeleteNote(ctx, entity.NoteUUID(noteUUID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := &pb.DeleteNoteResponse{}

	return resp, nil
}

func (c *Controller) SubscribeToEvents(
	_ *pb.SubscribeToEventsRequest,
	stream pb.NotesAPI_SubscribeToEventsServer,
) error {
	c.lgr.Info("subscribing to events")
	defer c.lgr.Info("unsubscribing from events")

	heartbeatTicker := time.NewTicker(1 * time.Microsecond) // first response immediately
	defer heartbeatTicker.Stop()

	eventTicker := time.NewTicker(10 * time.Second)
	defer eventTicker.Stop()

	ctx := stream.Context()

	var noteNum int

	for {
		select {
		case <-ctx.Done():
			c.lgr.Info("stream context done")
			return nil
		case <-heartbeatTicker.C:
			heartbeatResp := &pb.SubscribeToEventsResponse_Heartbeat{
				Heartbeat: &pb.HeartbeatEvent{
					Timestamp: responseDateTime(time.Now()),
				},
			}
			resp := &pb.SubscribeToEventsResponse{Payload: heartbeatResp}
			err := stream.Send(resp)
			if err != nil {
				c.lgr.Error("send response", slog.String("error", err.Error()))
			}

			heartbeatTicker.Reset(5 * time.Second)
		case <-eventTicker.C:
			noteNum++
			noteUUID, err := uuid.NewV7()
			if err != nil {
				c.lgr.Error("create uuid", slog.String("error", err.Error()))
				continue
			}

			createNoteEventResp := &pb.SubscribeToEventsResponse_CreatedNote{
				CreatedNote: &pb.CreateNoteEvent{
					Uuid:  noteUUID.String(),
					Title: fmt.Sprintf("Note #%d", noteNum),
				},
			}

			resp := &pb.SubscribeToEventsResponse{Payload: createNoteEventResp}
			err = stream.Send(resp)
			if err != nil {
				c.lgr.Error("send response", slog.String("error", err.Error()))
			}
		}
	}
}

func (c *Controller) UploadMetrics(stream pb.NotesAPI_UploadMetricsServer) error {
	c.lgr.Info("start stream processing")
	defer c.lgr.Info("end stream processing")

	ctx := stream.Context()

	var metricsSum int64

	for {
		if ctx.Err() != nil {
			c.lgr.Info("stream context error", slog.String("error", ctx.Err().Error()))
			break
		}

		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.lgr.Info("client close stream")
				break
			}

			c.lgr.Error("receive message from client", slog.String("error", err.Error()))
			continue
		}

		c.lgr.Info("receive metric", slog.Int64("value", req.Value))
		metricsSum += req.Value
	}

	resp := &pb.UploadMetricsResponse{
		Sum: metricsSum,
	}
	err := stream.SendAndClose(resp)
	if err != nil {
		return fmt.Errorf("send response: %w", err)
	}

	return nil
}
