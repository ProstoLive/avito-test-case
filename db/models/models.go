package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type User struct {
	UserID   string `db:"user_id"`
	Username string `db:"username"`
	TeamName string `db:"team_name"`
	IsActive bool   `db:"is_active"`
}

type ScannableUsers []User

type Team struct {
	TeamName string         `db:"team_name" json:"team_name"`
	Members  ScannableUsers `db:"members" json:"members"`
}

type TeamMember struct {
	UserID   string `db:"user_id"`
	Username string `db:"username"`
	IsActive bool   `db:"is_active"`
}

type UserPullRequest struct {
	PullRequestID     string         `db:"pull_request_id"`
	PullRequestName   string         `db:"pull_request_name"`
	AuthorID          string         `db:"author_id"`
	Status            string         `db:"status"`
	AssignedReviewers pq.StringArray `db:"assigned_reviewers"`
}

type PullRequest struct {
	UserPullRequest
	CreatedAt *time.Time `db:"created_at" json:"createdAt,omitempty"`
	MergedAt  *time.Time `db:"merged_at" json:"mergedAt,omitempty"`
}

type RequestPullRequestMerge struct {
	PullRequestId string `db:"pull_request_id" json:"pull_request_id"`
}

type ResponsePullRequest struct {
	PR PullRequest `json:"pr"`
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
