package database

import (
	"context"

	"go.uber.org/zap"

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

type Upsert struct {
	Key                string
	Value              interface{}
	ContentType        string
	PreserveTimestamps bool
	Tags               []Tag
}

type BatchInsert []Insert
type BatchUpsert []Upsert
type BatchDeleteByKey []string

type ExecResult struct {
	RowsAffected uint64
}

type Engine interface {
	BatchInsert(ctx context.Context, dbName string, bi BatchInsert) (*ExecResult, error)
	BatchUpsert(ctx context.Context, dbName string, bu BatchUpsert) (*ExecResult, error)
	BatchDeleteByKey(ctx context.Context, dbName string, keys BatchDeleteByKey) (*ExecResult, error)
	MGet(ctx context.Context, database string, keys []string) (map[string]*lemon.Document, error)
}

// LemonEngine wraps and manages the database store
type LemonEngine struct {
	store *Store
	lg    *zap.SugaredLogger
}

// NewEngine - creates a new LemonEngine
func NewEngine(store *Store, lg *zap.SugaredLogger) *LemonEngine {
	return &LemonEngine{
		store: store,
		lg:    lg,
	}
}

func (le *LemonEngine) MGet(ctx context.Context, database string, keys []string) (map[string]*lemon.Document, error) {
	db, err := le.store.Get(database)
	if err != nil {
		return nil, err
	}

	// todo: add MGetContext to LemonDB
	documentMap, err := db.MGetContext(ctx, keys...)
	if err != nil {
		return nil, errors.Wrap(ErrEngineFailed, err.Error())
	}

	return documentMap, nil
}

func (le *LemonEngine) BatchInsert(ctx context.Context, dbName string, bi BatchInsert) (*ExecResult, error) {
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

	return &ExecResult{
		RowsAffected: uint64(len(bi)),
	}, nil
}

func (le *LemonEngine) BatchDeleteByKey(
	ctx context.Context,
	dbName string,
	keys BatchDeleteByKey,
) (*ExecResult, error) {
	db, err := le.store.Get(dbName)
	if err != nil {
		return nil, err
	}

	deleted := 0
	if err := db.Update(ctx, func(tx *lemon.Tx) error {
		for _, k := range keys {
			if err := ctx.Err(); err != nil {
				return err
			}

			if err := tx.Remove(k); err != nil {
				le.lg.Infof("could not find key '%s' to remove from database '%s'", k, dbName)
				continue
			} else {
				deleted++
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &ExecResult{
		RowsAffected: uint64(deleted),
	}, nil
}

func (le *LemonEngine) BatchUpsert(ctx context.Context, dbName string, bi BatchUpsert) (*ExecResult, error) {
	db, err := le.store.Get(dbName)
	if err != nil {
		return nil, err
	}

	if err := db.Update(ctx, func(tx *lemon.Tx) error {
		for i := range bi {
			metaAppliers := make([]lemon.MetaApplier, 0, 3)

			ct := lemon.ContentTypeIdentifier(bi[i].ContentType)
			if ct != "" {
				metaAppliers = append(metaAppliers, lemon.WithContentType(ct))
			}

			if bi[i].PreserveTimestamps {
				metaAppliers = append(metaAppliers, lemon.WithTimestamps())
			}

			if bi[i].Tags != nil {
				m := make(lemon.M)
				for _, tag := range bi[i].Tags {
					m[tag.Name] = tag.Value
				}
				metaAppliers = append(metaAppliers, lemon.WithTags().Map(m))
			}

			if err := tx.InsertOrReplace(
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

	return &ExecResult{
		RowsAffected: uint64(len(bi)),
	}, nil
}
