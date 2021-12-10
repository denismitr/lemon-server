package database

import (
	"context"
	"github.com/denismitr/lemon"
)

type Tag struct {
	Name  string
	Value interface{}
	Type  uint32
}

type Insert struct {
	Key            string
	Value          interface{}
	ContentType    string
	WithTimestamps bool
}

type BatchInsert struct {
	Inserts []Insert
}

type InsertResult struct {
	RowsAffected uint64
}

type Engine interface {
	BatchInsert(ctx context.Context, dbName string, bi *BatchInsert) (*InsertResult, error)
}

type LemonEngine struct {
	store *Store
}

func NewEngine(store *Store) Engine {
	return &LemonEngine{
		store: store,
	}
}

func (le *LemonEngine) BatchInsert(ctx context.Context, dbName string, bi *BatchInsert) (*InsertResult, error) {
	db, err := le.store.Get(dbName)
	if err != nil {
		return nil, err
	}

	if err := db.Update(ctx, func(tx *lemon.Tx) error {
		for i := range bi.Inserts {
			ct := lemon.ContentTypeIdentifier(bi.Inserts[i].ContentType)
			if err := tx.Insert(bi.Inserts[i].Key, bi.Inserts[i].Value, lemon.WithContentType(ct)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &InsertResult{
		RowsAffected: uint64(len(bi.Inserts)),
	}, nil
}
