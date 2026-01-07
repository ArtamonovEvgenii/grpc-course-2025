package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"buf.build/go/protovalidate"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
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
		return nil, domainToTransportError(codes.Internal, err)
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
	c.lgr.Info("metrics start stream processing")
	defer c.lgr.Info("metrics end stream processing")

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

func (c *Controller) Chat(stream pb.NotesAPI_ChatServer) error {
	c.lgr.Info("chat start stream processing")
	defer c.lgr.Info("chat end stream processing")

	messagesCh := make(chan entity.ChatMessage, 10)
	defer close(messagesCh)

	ctx := stream.Context()
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(c.receiveChatMessages(egCtx, stream, messagesCh))
	eg.Go(c.responseChatMessages(egCtx, stream, messagesCh))
	eg.Go(c.heartbeatChatMessages(egCtx, stream))
	err := eg.Wait()
	if err != nil {
		return fmt.Errorf("err group wait: %w", err)
	}

	return nil
}

func (c *Controller) receiveChatMessages(
	ctx context.Context,
	stream pb.NotesAPI_ChatServer,
	ch chan<- entity.ChatMessage,
) func() error {
	return func() error {
		c.lgr.Info("receive chat messages processing start")
		defer c.lgr.Info("receive chat messages processing stop")

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

			c.lgr.Info("receive message", slog.String("correlation_id", req.CorrelationId))

			msg := entity.ChatMessage{
				CorrelationID: req.CorrelationId,
				Text:          req.Text,
			}

			select {
			case ch <- msg:
				c.lgr.Debug("put message to channel", slog.String("correlation_id", req.CorrelationId))
			default:
				c.lgr.Error("put message to channel", slog.String("correlation_id", req.CorrelationId))
			}
		}

		return nil
	}
}

func (c *Controller) responseChatMessages(
	ctx context.Context,
	stream pb.NotesAPI_ChatServer,
	ch <-chan entity.ChatMessage,
) func() error {
	return func() error {
		c.lgr.Info("response chat messages processing start")
		defer c.lgr.Info("response chat messages processing stop")

		for {
			select {
			case <-ctx.Done():
				return nil
			case msg, ok := <-ch:
				if !ok {
					return fmt.Errorf("message channel closed")
				}

				time.Sleep(1 * time.Second)

				var respMsg *pb.ChatMessageResponse
				if strings.Contains(msg.Text, "error") {
					respMsg = &pb.ChatMessageResponse{
						Payload: &pb.ChatMessageResponse_Error{Error: &pb.ChatMessageErrorResponse{
							CorrelationId: msg.CorrelationID,
							Error:         fmt.Sprintf("some error happens for: %s", msg.Text),
						}},
					}
				} else {
					respMsg = &pb.ChatMessageResponse{
						Payload: &pb.ChatMessageResponse_Success{Success: &pb.ChatMessageSuccessResponse{
							CorrelationId: msg.CorrelationID,
							Text:          fmt.Sprintf("response for: %s", msg.Text),
						}},
					}
				}

				err := stream.Send(respMsg)
				if err != nil {
					c.lgr.Error("send response", slog.String("error", err.Error()))
				}
			}
		}
	}
}

func (c *Controller) heartbeatChatMessages(
	ctx context.Context,
	stream pb.NotesAPI_ChatServer,
) func() error {
	return func() error {
		c.lgr.Info("heartbeat chat messages processing start")
		defer c.lgr.Info("heartbeat chat messages processing stop")

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		var heartbeatNum int

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				heartbeatNum++

				heartbeatMsg := &pb.ChatMessageResponse{
					Payload: &pb.ChatMessageResponse_Success{Success: &pb.ChatMessageSuccessResponse{
						CorrelationId: "none",
						Text:          fmt.Sprintf("heartbeat: %d", heartbeatNum),
					}},
				}

				err := stream.Send(heartbeatMsg)
				if err != nil {
					c.lgr.Error("send heartbeat message", slog.String("error", err.Error()))
				}
			}
		}
	}
}
