package domain

import "time"

type Rating string

const (
	RatingExplicit     Rating = "e"
	RatingSensitive    Rating = "s"
	RatingGeneral      Rating = "g"
	RatingQuestionable Rating = "q"
)

type Image struct {
	ID          string     `json:"id"`
	Upstream    SourceCode `json:"upstream"`
	CreatedAt   time.Time  `json:"created_at"`
	Source      *string    `json:"source"`
	Rating      Rating     `json:"rating"`
	Tags        []string   `json:"tags,omitempty"`
	HasChildren bool       `json:"has_children"`
	ParentID    *string    `json:"parent_id"`
	MD5         string     `json:"md5"`
	PreviewURL  string     `json:"preview_url,omitempty"`
	SampleURL   string     `json:"sample_url,omitempty"`
	FileURL     string     `json:"file_url"`
}
