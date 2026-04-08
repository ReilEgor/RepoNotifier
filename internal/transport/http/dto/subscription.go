package dto

import "time"

// SubscriptionResponse represents a user's subscription record.
type SubscriptionResponse struct {
	ID           int64     `json:"id" example:"1"`                            // Unique identifier of the subscription.
	Email        string    `json:"email" example:"user@example.com"`          // Subscriber's email address.
	RepositoryID int64     `json:"repository_id" example:"101"`               // Reference to the repository ID.
	CreatedAt    time.Time `json:"created_at" example:"2026-04-08T18:00:00Z"` // Timestamp when the subscription was created.
}

// CreateSubscriptionRequest defines the input payload for creating a new subscription.
type CreateSubscriptionRequest struct {
	Email      string `json:"email" binding:"required,email" example:"user@example.com"` // User email address for notifications.
	Repository string `json:"repository" binding:"required" example:"golang/go"`         // GitHub repository full name (owner/repo).
}
type DeleteSubscriptionRequest struct {
	Email      string `json:"email" binding:"required,email" example:"user@example.com"` // Email address associated with the subscription.
	Repository string `json:"repository" binding:"required" example:"golang/go"`         // Repository name to unsubscribe from.
}
