package dto

import "time"

type ReleaseInfoResponse struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
	URL         string    `json:"url"`
}
