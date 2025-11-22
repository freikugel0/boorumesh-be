package service

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/freikugel0/boorumesh-be/internal/domain"
	"github.com/freikugel0/boorumesh-be/internal/repository"
)

var ErrSourceNotFound = errors.New("source not found")

type DevSourceService interface {
	CreateSource(ctx context.Context, in CreateSourceInput) (domain.Source, error)
	GetSourceByCode(ctx context.Context, code string) (domain.Source, error)
}

type CreateSourceInput struct {
	Code     string                `json:"code"`
	Name     string                `json:"name"`
	BaseURL  string                `json:"base_url"`
	Enabled  *bool                 `json:"enabled,omitempty"`
	Request  domain.RequestConfig  `json:"request"`
	Mapping  domain.SourceMapping  `json:"mapping"`
	Defaults domain.SourceDefaults `json:"defaults"`
}

type devSourceService struct {
	repo repository.SourceRepository
}

func NewDevSourceService(repo repository.SourceRepository) DevSourceService {
	return &devSourceService{repo: repo}
}

func (s *devSourceService) CreateSource(ctx context.Context, in CreateSourceInput) (domain.Source, error) {
	code := strings.TrimSpace(in.Code)
	name := strings.TrimSpace(in.Name)
	base := strings.TrimSpace(in.BaseURL)

	if code == "" {
		return domain.Source{}, errors.New("code is required")
	}
	if name == "" {
		return domain.Source{}, errors.New("name is required")
	}
	if base == "" {
		return domain.Source{}, errors.New("base_url is required")
	}
	if _, err := url.ParseRequestURI(base); err != nil {
		return domain.Source{}, errors.New("base_url invalid")
	}
	if strings.TrimSpace(in.Request.PostsPath) == "" {
		return domain.Source{}, errors.New("request.posts_path is required")
	}
	if in.Mapping.Fields == nil ||
		in.Mapping.Fields["id"].Key == "" ||
		in.Mapping.Fields["file_url"].Key == "" {
		return domain.Source{}, errors.New("mapping.fields must include at least 'id' and 'file_url'")
	}

	req := in.Request
	if req.TagsParam == "" {
		req.TagsParam = "tags"
	}
	if req.LimitParam == "" {
		req.LimitParam = "limit"
	}
	if req.PageParam == "" {
		req.PageParam = "page"
	}
	if req.Headers == nil {
		req.Headers = map[string]string{}
	}
	if _, ok := req.Headers["User-Agent"]; !ok {
		req.Headers["User-Agent"] = "boorumesh/1.0"
	}

	def := in.Defaults
	if def.MaxLimit == 0 {
		def.MaxLimit = 100
	}
	if def.TimeoutMS == 0 {
		def.TimeoutMS = 5000
	}

	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}

	src := domain.Source{
		Code:     domain.SourceCode(code),
		Name:     name,
		BaseURL:  strings.TrimRight(base, "/"),
		Enabled:  enabled,
		Request:  req,
		Mapping:  in.Mapping,
		Defaults: def,
	}

	out, err := s.repo.Create(ctx, src)
	if err != nil {
		return domain.Source{}, err
	}
	return out, nil
}

func (s *devSourceService) GetSourceByCode(ctx context.Context, code string) (domain.Source, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.Source{}, errors.New("code is required")
	}

	src, err := s.repo.GetByCode(ctx, domain.SourceCode(code))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.Source{}, ErrSourceNotFound
		}
		return domain.Source{}, err
	}

	return src, nil
}
