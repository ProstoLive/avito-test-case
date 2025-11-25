package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	UserID   string `db:"user_id"`
	Username string `db:"username"`
	TeamName string `db:"team_name"`
	IsActive bool   `db:"is_active"`
}

type ScannableUsers []User

type Team struct {
	TeamName string `db:"team_name" json:"team_name"`
	Members  ScannableUsers `db:"members" json:"members"`
}

type TeamMember struct {
	UserID   string `db:"user_id"`
	Username string `db:"username"`
	IsActive bool   `db:"is_active"`
}

type PullRequest struct {
	PullRequestID     string    `db:"pull_request_id"`
	PullRequestName   string    `db:"pull_request_name"`
	AuthorID          string    `db:"author_id"`
	Status            string    `db:"status"`
	AssignedReviewers []string  `db:"assigned_reviewers"`
	CreatedAt         time.Time `db:"createdAt,omitempty"`
	MergedAt          time.Time `db:"merged_at,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   string `db:"pull_request_id"`
	PullRequestName string `db:"pull_request_name"`
	AuthorID        string `db:"author_id"`
	Status          string `db:"status"`
}

func (u *ScannableUsers) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Users: expected []byte, got %T", src)
	}
	return json.Unmarshal(bytes, u)
}
