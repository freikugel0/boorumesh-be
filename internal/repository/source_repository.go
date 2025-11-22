package repository

import (
	"context"

	"github.com/freikugel0/boorumesh-be/internal/domain"
)

var (
	ErrSourceExists = fmtError("source already exists")
	ErrNotFound     = fmtError("not found")
)

type fmtError string

func (e fmtError) Error() string { return string(e) }

type SourceRepository interface {
	Create(ctx context.Context, src domain.Source) (domain.Source, error)
	GetByCode(ctx context.Context, code domain.SourceCode) (domain.Source, error)
}
