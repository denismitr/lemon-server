package httpd

import "github.com/denismitr/lemon-server/command"

type Database interface {
	BatchInsert(er *command.BatchInsertRequest) ([]*command.ExecuteResult, error)
	BatchUpsert(er *command.BatchUpsertRequest) ([]*command.ExecuteResult, error)
	BatchDelete(er *command.BatchDeleteByKeyRequest) ([]*command.ExecuteResult, error)
}

type DBResult struct {
	ExecuteResult []*command.ExecuteResult
}
