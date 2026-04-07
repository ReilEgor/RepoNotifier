package model

import "time"

type Subscription struct {
	ID           int64
	Email        string
	RepositoryID int64
	CreatedAt    time.Time
}
