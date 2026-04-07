package dto

type RepositoryResponse struct {
	ID          int64  `json:"id"`
	FullName    string `json:"full_name"`
	LastSeenTag string `json:"last_seen_tag"`
}
