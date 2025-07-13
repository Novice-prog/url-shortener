package service

import (
	"context"
	"errors"
	"url-shortener1/internal/model"
	"url-shortener1/internal/repository"
	"url-shortener1/pkg/shortener"
)

var (
	ErrNotFound = errors.New("url not found")      // 404
	ErrExists   = errors.New("url already exists") // 409
)

type URLService struct {
	repo  repository.URLRepository
	idLen int
}

func NewURLService(r repository.URLRepository, idLen int) *URLService {
	return &URLService{repo: r, idLen: idLen}
}

func (s *URLService) Create(ctx context.Context, original string) (*model.Shorten, error) {
	sh := &model.Shorten{OriginalURL: original}
	for {
		sh.ShortURL = shortener.RandomURL(s.idLen)

		if err := s.repo.Save(ctx, sh); err != nil {
			if errors.Is(err, model.ErrURLExists) {
				continue
			}
			return nil, err
		}
		return sh, nil
	}
}

func (s *URLService) Resolve(ctx context.Context, short string) (string, error) {
	sh, err := s.repo.Find(ctx, short)

	if err != nil {
		if errors.Is(err, model.ErrURLNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	_ = s.repo.IncVisits(ctx, short)

	return sh.OriginalURL, nil
}

func (s *URLService) GetStat(ctx context.Context, short string) (*model.Shorten, error) {
	sh, err := s.repo.Find(ctx, short)
	if err != nil {
		if errors.Is(err, model.ErrURLNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return sh, nil
}
