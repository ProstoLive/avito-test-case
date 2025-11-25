package dto

type UserSetIsActive struct {
	UserID   string `db:"user_id" json:"user_id"`
	IsActive bool   `db:"is_active" json:"is_active"`
}

type CreatePR struct {
	PRID     string `db:"pull_request_id" json:"pull_request_id"`
	PRName   string `db:"pull_request_name" json:"pull_request_name"`
	AuthorID string `db:"author_id" json:"author_id"`
}
