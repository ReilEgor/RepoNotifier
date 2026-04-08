package model

import "time"

type Subscription struct {
	ID           int64
	UserID       int64
	RepositoryID int64
	CreatedAt    time.Time
}
