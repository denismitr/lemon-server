package serverpb

import (
	"context"
	"fmt"
	"time"

	"github.com/denismitr/lemon-server/internal/database"
	"github.com/denismitr/lemon-server/pkg/command"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandlers struct {
	lg *zap.SugaredLogger
	db database.Engine
	//command.UnimplementedReceiverServer
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

	bi, err := database.ConvertGrpcToLemonInsert(request)
	if err != nil {
		grpcErr := createBatchInsertGrpcError(err)
		g.lg.Error(err)
		return nil, grpcErr
	}

	ir, err := g.db.BatchInsert(ctx, request.Database, bi)
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

func (g *GrpcHandlers) MGet(
	ctx context.Context,
	request *command.MultiGetQueryRequest,
) (*command.QueryResult, error) {
	start := time.Now()

	documents, err := g.db.MGet(ctx, request.Database, request.Keys)
	if err != nil {
		errorStatus := status.New(codes.Internal, err.Error())
		return nil, errorStatus.Err()
	}

	if !request.IgnoreMissing && len(request.Keys) != len(documents) {
		errorStatus := status.New(codes.NotFound, "some keys are missing, cannot ignore missing")
		ds, err := errorStatus.WithDetails(
			&errdetails.ErrorInfo{
				Reason: "request required not to ignore missing keys",
				Metadata: map[string]string{
					"expected": fmt.Sprintf("%d keys", len(request.Keys)),
					"got":      fmt.Sprintf("%d keys", len(documents)),
				},
			},
		)

		if err != nil {
			return nil, errorStatus.Err()
		}

		g.lg.Error(ds.Err())
		return nil, ds.Err()
	}

	result := command.QueryResult{
		Documents: make(map[string]*command.Document, len(documents)),
	}

	for key, document := range documents {
		grpcDoc, err := database.ConvertLemonToGrpcDocument(document)
		if err != nil {
			g.lg.Error(err)
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		result.Documents[key] = grpcDoc
	}

	result.Elapsed = time.Since(start).Milliseconds()

	return &result, nil
}

func (g GrpcHandlers) PingPong(ctx context.Context, ping *command.Ping) (*command.Pong, error) {
	return &command.Pong{
		Message: "pong",
	}, nil
}
