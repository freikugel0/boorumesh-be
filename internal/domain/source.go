package domain

import "time"

type SourceCode string

type FieldMapping struct {
	Key   string `json:"key"`
	Split string `json:"split,omitempty"`
}

type SourceMapping struct {
	Fields map[string]FieldMapping `json:"fields"`
}

type RequestConfig struct {
	PostsPath  string            `json:"posts_path"`
	TagsParam  string            `json:"tags_param"`
	LimitParam string            `json:"limit_param"`
	PageParam  string            `json:"page_param"`
	ExtraQuery map[string]string `json:"extra_query,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
}

type SourceDefaults struct {
	TagsSuffix string `json:"tags_suffix,omitempty"`
	MaxLimit   int    `json:"max_limit,omitempty"`
	TimeoutMS  int    `json:"timeout_ms,omitempty"`
}

type Source struct {
	ID        int64          `json:"id"`
	Code      SourceCode     `json:"code"`
	Name      string         `json:"name"`
	BaseURL   string         `json:"base_url"`
	Enabled   bool           `json:"enabled"`
	Request   RequestConfig  `json:"request"`
	Mapping   SourceMapping  `json:"mapping"`
	Defaults  SourceDefaults `json:"defaults"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
