package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/freikugel0/boorumesh-be/internal/domain"
	"github.com/freikugel0/boorumesh-be/internal/repository"
)

var (
	ErrSourceDisabled = errors.New("source is disabled")
)

type SourceFetchService interface {
	FetchBySource(ctx context.Context, code string, tags []string, page, limit int) ([]domain.Image, error)
}

type sourceFetchService struct {
	repo       repository.SourceRepository
	httpClient *resty.Client
}

func NewSourceFetchService(repo repository.SourceRepository) SourceFetchService {
	client := resty.New().SetTimeout(10 * time.Second)

	return &sourceFetchService{
		repo:       repo,
		httpClient: client,
	}
}

func (s *sourceFetchService) FetchBySource(ctx context.Context, code string, tags []string, page, limit int) ([]domain.Image, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("code is required")
	}

	// Get source code
	src, err := s.repo.GetByCode(ctx, domain.SourceCode(code))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSourceNotFound
		}
		return nil, err
	}
	if !src.Enabled {
		return nil, ErrSourceDisabled
	}

	// Apply default limit / page
	if page <= 0 {
		page = 1
	}
	maxLimit := src.Defaults.MaxLimit
	if maxLimit <= 0 {
		maxLimit = 100
	}
	if limit <= 0 || limit > maxLimit {
		limit = maxLimit
	}

	// Build base URL: base_url + posts_path
	baseURL := strings.TrimRight(src.BaseURL, "/") + "/" + strings.TrimLeft(src.Request.PostsPath, "/")

	// Context with timeout per-source
	if src.Defaults.TimeoutMS > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(src.Defaults.TimeoutMS)*time.Millisecond)
		defer cancel()
	}

	// Build resty request
	req := s.httpClient.R().SetContext(ctx).SetHeaders(src.Request.Headers)

	// Query params
	if len(tags) > 0 {
		req.SetQueryParam(src.Request.LimitParam, strconv.Itoa(limit)).SetQueryParam(src.Request.PageParam, strconv.Itoa(page))
	}

	// Exec request
	resp, err := req.Get(baseURL)
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode())
	}

	// Decode JSON
	var rawPosts []map[string]any
	if err := json.Unmarshal(resp.Body(), &rawPosts); err != nil {
		return nil, fmt.Errorf("failed to decode upstream json: %w", err)
	}

	// Map response to domain.Image
	images := make([]domain.Image, 0, len(rawPosts))
	for _, raw := range rawPosts {
		img, err := mapRawToImage(src, raw)
		if err != nil {
			continue
		}
		images = append(images, img)
	}

	return images, nil
}

func mapRawToImage(src domain.Source, raw map[string]any) (domain.Image, error) {
	m := src.Mapping.Fields

	// Required fields

	idMapping, ok := m["id"]
	if !ok || idMapping.Key == "" {
		return domain.Image{}, errors.New("mapping for 'id' is missing")
	}
	fileMapping, ok := m["file_url"]
	if !ok || fileMapping.Key == "" {
		return domain.Image{}, errors.New("mapping for 'file_url' is missing")
	}

	getVal := func(key string) (any, bool) {
		v, ok := raw[key]
		return v, ok
	}

	getStr := func(key string) (string, bool) {
		v, ok := getVal(key)
		if !ok {
			return "", false
		}
		return fmt.Sprint(v), true
	}

	// id
	id, ok := getStr(idMapping.Key)
	if !ok {
		return domain.Image{}, fmt.Errorf("id key %q not found in upstream json", idMapping.Key)
	}

	// file_url
	fileURL, ok := getStr(fileMapping.Key)
	if !ok {
		return domain.Image{}, fmt.Errorf("file_url key %q not found in upstream json", fileMapping.Key)
	}

	// Optional fields

	// created_at
	var createdAt time.Time
	if createdMapping, ok := m["created_at"]; ok && createdMapping.Key != "" {
		if v, ok := getVal(createdMapping.Key); ok {
			if t, err := parseTimeFlexible(v); err == nil {
				createdAt = t
			}
		}
	}

	// image_src_url
	var imageSrc string
	if imgSrcMapping, ok := m["image_src_url"]; ok && imgSrcMapping.Key != "" {
		if s, ok := getStr(imgSrcMapping.Key); ok {
			imageSrc = s
		}
	}

	// rating (string → domain.Rating)
	var rating domain.Rating
	if ratingMapping, ok := m["rating"]; ok && ratingMapping.Key != "" {
		if s, ok := getStr(ratingMapping.Key); ok && s != "" {
			s = strings.ToLower(s)
			switch s {
			case "e", "explicit":
				rating = domain.RatingExplicit
			case "q", "questionable":
				rating = domain.RatingQuestionable
			case "s", "sensitive":
				rating = domain.RatingSensitive
			case "g", "general", "safe":
				rating = domain.RatingGeneral
			default:
				// kalau nggak kebaca, biarin kosong aja
			}
		}
	}

	// has_children (bool / "true"/"1")
	hasChildren := false
	if hcMapping, ok := m["has_children"]; ok && hcMapping.Key != "" {
		if v, ok := getVal(hcMapping.Key); ok {
			hasChildren = toBoolFlexible(v)
		}
	}

	// parent_id
	var parentID string
	if parentMapping, ok := m["parent_id"]; ok && parentMapping.Key != "" {
		if s, ok := getStr(parentMapping.Key); ok {
			parentID = s
		}
	}

	// md5
	var md5 string
	if md5Mapping, ok := m["md5"]; ok && md5Mapping.Key != "" {
		if s, ok := getStr(md5Mapping.Key); ok {
			md5 = s
		}
	}

	// tags (string "a b c" atau array)
	var tags []string
	if tagsMapping, ok := m["tags"]; ok && tagsMapping.Key != "" {
		if v, ok := getVal(tagsMapping.Key); ok {
			switch t := v.(type) {
			case string:
				if t != "" {
					tags = strings.Fields(t)
				}
			case []any:
				for _, tv := range t {
					tags = append(tags, fmt.Sprint(tv))
				}
			}
		}
	}

	// preview_url
	var preview string
	if previewMapping, ok := m["preview_url"]; ok && previewMapping.Key != "" {
		if s, ok := getStr(previewMapping.Key); ok {
			preview = s
		}
	}

	// sample_url
	var sample string
	if sampleMapping, ok := m["sample_url"]; ok && sampleMapping.Key != "" {
		if s, ok := getStr(sampleMapping.Key); ok {
			sample = s
		}
	}

	return domain.Image{
		ID:          id,
		Source:      src.Code,
		CreatedAt:   createdAt,
		ImageSrcURL: imageSrc,
		Rating:      rating,
		Tags:        tags,
		HasChildren: hasChildren,
		ParentID:    parentID,
		MD5:         md5,
		PreviewURL:  preview,
		SampleURL:   sample,
		FileURL:     fileURL,
	}, nil
}

func parseTimeFlexible(v any) (time.Time, error) {
	switch t := v.(type) {
	case string:
		t = strings.TrimSpace(t)
		if t == "" {
			return time.Time{}, fmt.Errorf("empty time string")
		}
		// try RFC3339 format (2023-10-01T12:34:56Z)
		if ts, err := time.Parse(time.RFC3339, t); err == nil {
			return ts, nil
		}
		// try unix format
		if unix, err := strconv.ParseInt(t, 10, 64); err == nil {
			return time.Unix(unix, 0).UTC(), nil
		}
		return time.Time{}, fmt.Errorf("unsupported time string %q", t)
	case float64:
		// typical JSON number → float64
		return time.Unix(int64(t), 0).UTC(), nil
	case int64:
		return time.Unix(t, 0).UTC(), nil
	case int:
		return time.Unix(int64(t), 0).UTC(), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time type %T", v)
	}
}

func toBoolFlexible(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		return s == "true" || s == "1" || s == "yes"
	case float64:
		return t != 0
	case int:
		return t != 0
	case int64:
		return t != 0
	default:
		return false
	}
}
