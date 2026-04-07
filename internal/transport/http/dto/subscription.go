package dto

import "time"

type SubscriptionResponse struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	RepositoryID int64     `json:"repository_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateSubscriptionRequest struct {
	Email      string `json:"email"`
	Repository string `json:"repository"`
}
type DeleteSubscriptionRequest struct {
	Email      string `json:"email"`
	Repository string `json:"repository"`
}
