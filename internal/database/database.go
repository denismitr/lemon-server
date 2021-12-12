package database

import (
	"context"
	"github.com/denismitr/lemon"
	"github.com/pkg/errors"
)

var ErrEngineFailed = errors.New("database engine faied")

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
	Tags           []Tag
}

type BatchInsert []Insert

type InsertResult struct {
	RowsAffected uint64
}

type Engine interface {
	BatchInsert(ctx context.Context, dbName string, bi BatchInsert) (*InsertResult, error)
	MGet(ctx context.Context, database string, keys []string) (map[string]*lemon.Document, error)
}

type LemonEngine struct {
	store *Store
}

func (le *LemonEngine) MGet(ctx context.Context, database string, keys []string) (map[string]*lemon.Document, error) {
	db, err := le.store.Get(database)
	if err != nil {
		return nil, err
	}

	// todo: add MGetContext to LemonDB
	documentMap, err := db.MGet(keys...)
	if err != nil {
		return nil, errors.Wrap(ErrEngineFailed, err.Error())
	}

	return documentMap, nil
}

func NewEngine(store *Store) Engine {
	return &LemonEngine{
		store: store,
	}
}

func (le *LemonEngine) BatchInsert(ctx context.Context, dbName string, bi BatchInsert) (*InsertResult, error) {
	db, err := le.store.Get(dbName)
	if err != nil {
		return nil, err
	}

	if err := db.Update(ctx, func(tx *lemon.Tx) error {
		for i := range bi {
			metaAppliers := make([]lemon.MetaApplier, 0, 2)

			ct := lemon.ContentTypeIdentifier(bi[i].ContentType)
			if ct != "" {
				metaAppliers = append(metaAppliers, lemon.WithContentType(ct))
			}

			if bi[i].Tags != nil {
				m := make(lemon.M)
				for _, tag := range bi[i].Tags {
					m[tag.Name] = tag.Value
				}
				metaAppliers = append(metaAppliers, lemon.WithTags().Map(m))
			}

			if err := tx.Insert(
				bi[i].Key,
				bi[i].Value,
				metaAppliers...,
			); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &InsertResult{
		RowsAffected: uint64(len(bi)),
	}, nil
}
