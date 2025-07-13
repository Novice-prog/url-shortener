package repository

import (
	"context"
	"errors"
	"url-shortener1/internal/model"
	"url-shortener1/internal/storage/sqlite"
)

type URLRepository interface {
	Save(ctx context.Context, sh *model.Shorten) error
	Find(ctx context.Context, short string) (*model.Shorten, error)
	IncVisits(ctx context.Context, short string) error
	Close() error
}

type SQLiteRepository struct {
	storage *sqlite.Storage
}

func NewSQLiteRepository(storage *sqlite.Storage) *SQLiteRepository {
	return &SQLiteRepository{storage: storage}
}

func (r *SQLiteRepository) Save(_ context.Context, sh *model.Shorten) error {
	if err := r.storage.Save(sh); err != nil {
		if errors.Is(err, sqlite.ErrUnique) {
			return model.ErrURLExists
		}
		return err
	}
	return nil
}

func (r *SQLiteRepository) Find(_ context.Context, short string) (*model.Shorten, error) {
	sh, err := r.storage.GetStats(short)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return nil, model.ErrURLNotFound
		}
		return nil, err
	}
	return sh, nil
}

func (r *SQLiteRepository) IncVisits(_ context.Context, short string) error {
	err := r.storage.IncVisits(short)
	if err != nil {
		if errors.Is(err, sqlite.ErrNoRowsUpdated) || errors.Is(err, sqlite.ErrNotFound) {
			return model.ErrURLNotFound
		}
		return err
	}
	return nil
}

func (r *SQLiteRepository) Close() error {
	return r.storage.Close()
}
