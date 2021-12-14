package database

import (
	"github.com/denismitr/lemon"
	"github.com/denismitr/lemon-server/pkg/command"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrInvalidInput = errors.New("invalid user input")
var ErrInvalidDocumentValue = errors.New("invalid document value")
var ErrInvalidTagValue = errors.New("invalid tag value")

func ConvertGrpcToLemonInsert(request *command.BatchInsertRequest) (BatchInsert, error) {
	bi := make(BatchInsert, len(request.Stmt))
	for i, stmt := range request.Stmt {
		bi[i].Key = stmt.Key
		bi[i].WithTimestamps = stmt.WithTimestamps
		bi[i].ContentType = stmt.ContentType

		switch typedValue := stmt.Value.(type) {
		case *command.InsertStatement_Blob:
			bi[i].Value = typedValue.Blob
		case *command.InsertStatement_Bool:
			bi[i].Value = typedValue.Bool
		case *command.InsertStatement_Int:
			bi[i].Value = typedValue.Int
		case *command.InsertStatement_Str:
			bi[i].Value = typedValue.Str
		default:
			return nil, errors.Wrapf(ErrInvalidDocumentValue, "value type %T unsupported", typedValue)
		}

		if stmt.Tags != nil {
			bi[i].Tags = make([]Tag, len(stmt.Tags))
			for j, tag := range stmt.Tags {
				bi[i].Tags[j].Name = tag.Name
				switch typedTagValue := tag.Value.(type) {
				case *command.Tag_Int:
					bi[i].Tags[j].Value = typedTagValue.Int
				case *command.Tag_Float:
					bi[i].Tags[j].Value = typedTagValue.Float
				case *command.Tag_Str:
					bi[i].Tags[j].Value = typedTagValue.Str
				case *command.Tag_Bool:
					bi[i].Tags[j].Value = typedTagValue.Bool
				default:
					return nil, errors.Wrapf(ErrInvalidTagValue, "value type %T unsupported", typedTagValue)
				}
			}
		}
	}

	return bi, nil
}

func ConvertGrpcToLemonUpsert(request *command.BatchUpsertRequest) (BatchUpsert, error) {
	bi := make(BatchUpsert, len(request.Stmt))
	for i, stmt := range request.Stmt {
		bi[i].Key = stmt.Key
		bi[i].PreserveTimestamps = stmt.PreserveTimestamps
		bi[i].ContentType = stmt.ContentType

		switch typedValue := stmt.Value.(type) {
		case *command.UpsertStatement_Blob:
			bi[i].Value = typedValue.Blob
		case *command.UpsertStatement_Bool:
			bi[i].Value = typedValue.Bool
		case *command.UpsertStatement_Int:
			bi[i].Value = typedValue.Int
		case *command.UpsertStatement_Str:
			bi[i].Value = typedValue.Str
		default:
			return nil, errors.Wrapf(ErrInvalidDocumentValue, "value type %T unsupported", typedValue)
		}

		if stmt.Tags != nil {
			if err := covertTags(bi, i, stmt); err != nil {
				return nil, err
			}
		}
	}

	return bi, nil
}

func covertTags(bi BatchUpsert, i int, stmt *command.UpsertStatement) error {
	bi[i].Tags = make([]Tag, len(stmt.Tags))
	for j, tag := range stmt.Tags {
		bi[i].Tags[j].Name = tag.Name
		switch typedTagValue := tag.Value.(type) {
		case *command.Tag_Int:
			bi[i].Tags[j].Value = typedTagValue.Int
		case *command.Tag_Float:
			bi[i].Tags[j].Value = typedTagValue.Float
		case *command.Tag_Str:
			bi[i].Tags[j].Value = typedTagValue.Str
		case *command.Tag_Bool:
			bi[i].Tags[j].Value = typedTagValue.Bool
		default:
			return errors.Wrapf(ErrInvalidTagValue, "value type %T unsupported", typedTagValue)
		}
	}
	return nil
}

func ConvertLemonToGrpcDocument(d *lemon.Document) (*command.Document, error) {
	var result command.Document

	result.Key = d.Key()
	result.CreatedAt = timestamppb.New(d.CreatedAt())
	result.UpdatedAt = timestamppb.New(d.UpdatedAt())
	result.ContentType = string(d.ContentType())
	result.Value = d.Value()

	for name, v := range d.Tags() {
		ct := &command.Tag{Name: name}
		switch typedTagValue := v.(type) {
		case int:
			ct.Value = &command.Tag_Int{Int: int64(typedTagValue)}
		case float64:
			ct.Value = &command.Tag_Float{Float: float64(typedTagValue)}
		case bool:
			ct.Value = &command.Tag_Bool{Bool: typedTagValue}
		case string:
			ct.Value = &command.Tag_Str{Str: typedTagValue}
		default:
			return nil, errors.Wrapf(ErrInvalidTagValue, "tag type %T unsupported", typedTagValue)
		}

		result.Tags = append(result.Tags, ct)
	}

	return &result, nil
}
