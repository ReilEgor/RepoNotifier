package model

import "time"

type ReleaseInfo struct {
	TagName     string
	PublishedAt time.Time
	URL         string
}
