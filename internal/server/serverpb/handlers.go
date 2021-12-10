package serverpb

import (
	"context"
	"github.com/denismitr/lemon-server/internal/database"
	"github.com/denismitr/lemon-server/pkg/command"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type GrpcHandlers struct {
	lg *zap.SugaredLogger
	db database.Engine
}

func NewHandlers(lg *zap.SugaredLogger, db database.Engine) *GrpcHandlers {
	return &GrpcHandlers{
		lg: lg,
		db: db,
	}
}

func (g GrpcHandlers) BatchUpsert(
	ctx context.Context,
	request *command.BatchUpsertRequest,
) (*command.ExecuteResult, error) {
	panic("implement")
}

func (g *GrpcHandlers) BatchInsert(
	ctx context.Context,
	request *command.BatchInsertRequest,
) (*command.ExecuteResult, error) {
	start := time.Now()

	bi := database.BatchInsert{
		Inserts: make([]database.Insert, len(request.Stmt)),
	}

	for i := range request.Stmt {
		bi.Inserts[i].Key = request.Stmt[i].Key
		bi.Inserts[i].WithTimestamps = request.Stmt[i].WithTimestamps
		bi.Inserts[i].ContentType = request.Stmt[i].ContentType

		switch typedValue := request.Stmt[i].Value.(type) {
		case *command.InsertStatement_Blob:
			bi.Inserts[i].Value = typedValue.Blob
		case *command.InsertStatement_Bool:
			bi.Inserts[i].Value = typedValue.Bool
		case *command.InsertStatement_Int:
			bi.Inserts[i].Value = typedValue.Int
		case *command.InsertStatement_Str:
			bi.Inserts[i].Value = typedValue.Str
		default:
			errorStatus := status.New(codes.InvalidArgument, "invalid insert value type")
			ds, err := errorStatus.WithDetails(
				&errdetails.BadRequest_FieldViolation{
					Field:       "Value",
					Description: "Must be of type string, blob, int or bool",
				},
			)

			if err != nil {
				return nil, errorStatus.Err()
			}

			g.lg.Error(ds.Err())
			return nil, ds.Err()
		}
	}

	ir, err := g.db.BatchInsert(ctx, request.Database, &bi)
	if err != nil {
		// todo: handle key already exists
		errorStatus := status.New(codes.Internal, err.Error())
		return nil, errorStatus.Err()
	}

	return &command.ExecuteResult{
		DocumentsAffected: ir.RowsAffected,
		Elapsed:           time.Since(start).Milliseconds(),
	}, nil
}

func (g GrpcHandlers) BatchDeleteByKey(ctx context.Context, request *command.BatchDeleteByKeyRequest) (*command.ExecuteResult, error) {
	panic("implement me")
}

func (g GrpcHandlers) MGet(ctx context.Context, request *command.MultiGetQueryRequest) (*command.QueryResult, error) {
	panic("implement me")
}

func (g GrpcHandlers) PingPong(ctx context.Context, ping *command.Ping) (*command.Pong, error) {
	return &command.Pong{
		Message: "pong",
	}, nil
}
