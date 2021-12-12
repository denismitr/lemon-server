package serverpb

import (
	"github.com/denismitr/lemon-server/internal/database"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createBatchInsertGrpcError(err error) error {
	if errors.Is(err, database.ErrInvalidDocumentValue) {
		errorStatus := status.New(codes.InvalidArgument, "invalid value type")
		ds, err := errorStatus.WithDetails(
			&errdetails.BadRequest_FieldViolation{
				Field:       "Value",
				Description: "Must be of type string, bytes, int64 or bool",
			},
		)

		if err != nil {
			return errorStatus.Err()
		}

		return ds.Err()
	}

	if errors.Is(err, database.ErrInvalidTagValue) {
		errorStatus := status.New(codes.InvalidArgument, "invalid tag value type")
		ds, err := errorStatus.WithDetails(
			&errdetails.BadRequest_FieldViolation{
				Field:       "Tags.Value",
				Description: "Must be of type string, float64, int64 or bool",
			},
		)

		if err != nil {
			return errorStatus.Err()
		}

		return ds.Err()
	}

	return err
}
