package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jackc/pgconn"

	"github.com/freikugel0/boorumesh-be/internal/domain"
	"github.com/freikugel0/boorumesh-be/internal/repository"
)

type SourceRepositoryPostgres struct {
	db *sql.DB
}

func NewSourceRepositoryPostgres(db *sql.DB) *SourceRepositoryPostgres {
	return &SourceRepositoryPostgres{db: db}
}

func (r *SourceRepositoryPostgres) Create(ctx context.Context, src domain.Source) (domain.Source, error) {
	reqJSON, err := json.Marshal(src.Request)
	if err != nil {
		return domain.Source{}, err
	}
	mapJSON, err := json.Marshal(src.Mapping)
	if err != nil {
		return domain.Source{}, err
	}
	defJSON, err := json.Marshal(src.Defaults)
	if err != nil {
		return domain.Source{}, err
	}

	const q = `
INSERT INTO sources (code, name, base_url, enabled, request, mapping, defaults)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at;
`

	row := r.db.QueryRowContext(ctx, q,
		src.Code,
		src.Name,
		src.BaseURL,
		src.Enabled,
		reqJSON,
		mapJSON,
		defJSON,
	)

	if err := row.Scan(&src.ID, &src.CreatedAt, &src.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Source{}, repository.ErrSourceExists
		}
		return domain.Source{}, err
	}

	return src, nil
}

func (r *SourceRepositoryPostgres) GetByCode(ctx context.Context, code domain.SourceCode) (domain.Source, error) {
	var (
		src     domain.Source
		reqJSON []byte
		mapJSON []byte
		defJSON []byte
	)

	const q = `
SELECT id, code, name, base_url, enabled, request, mapping, defaults, created_at, updated_at
FROM sources
WHERE code = $1;
`

	err := r.db.QueryRowContext(ctx, q, code).Scan(
		&src.ID,
		&src.Code,
		&src.Name,
		&src.BaseURL,
		&src.Enabled,
		&reqJSON,
		&mapJSON,
		&defJSON,
		&src.CreatedAt,
		&src.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Source{}, repository.ErrNotFound
		}
		return domain.Source{}, err
	}

	if err := json.Unmarshal(reqJSON, &src.Request); err != nil {
		return domain.Source{}, err
	}
	if err := json.Unmarshal(mapJSON, &src.Mapping); err != nil {
		return domain.Source{}, err
	}
	if err := json.Unmarshal(defJSON, &src.Defaults); err != nil {
		return domain.Source{}, err
	}

	return src, nil
}
