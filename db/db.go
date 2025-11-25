package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"prmanagement/api/dto"
	"prmanagement/db/models"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var DB *sqlx.DB

func Connect() error {

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = sqlx.Connect("postgres", connStr)
	return err
}

func AddTeam(newTeam *models.Team) error {
	var existTeam models.Team

	err := DB.Get(
		&existTeam,
		`SELECT team_name FROM teams 
		 WHERE team_name=$1`,
		newTeam.TeamName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonMembers, err := json.Marshal(newTeam.Members)
			if err != nil {
				fmt.Printf("Failed marshalling members array, %v\n", err)
				return nil
			}

			_, err = DB.Exec(
				`INSERT INTO teams (team_name, members)
				 VALUES ($1, $2)`,
				newTeam.TeamName,
				jsonMembers,
			)

			if err != nil {
				fmt.Printf("Error inserting new team, %v\n", err)
				return nil
			} else {
				fmt.Printf("New team %s has been registered\n", newTeam.TeamName)
			}
		}
	} else {
		return errors.New("TEAM_EXISTS")
	}
	return nil
}

func GetTeam(teamName string) (*models.Team, error) {
	var resultTeam models.Team
	err := DB.Get(
		&resultTeam,
		`SELECT team_name, members FROM teams 
			WHERE team_name=$1`,
		teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("NOT_FOUND")
		} else {
			fmt.Printf("Failed to get team from db, error: %v", err)
			return nil, err
		}
	}
	return &resultTeam, nil
}

func UserSetIsActive(newStatus *dto.UserSetIsActive) (*models.User, error) {
	var existUser models.User
	err := DB.Get(&existUser, "SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1", newStatus.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("NOT_FOUND")
		} else {
			fmt.Printf("Failed search for user in database. Error: %v", err)
			return nil, err
		}
	}

	_, err = DB.Exec("UPDATE users SET is_active = $1 WHERE user_id = $2", newStatus.IsActive, newStatus.UserID)

	if err != nil {
		fmt.Printf("Failed update status in database, error: %v", err)
		return nil, err
	}

	err = DB.Get(&existUser, "SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1", newStatus.UserID)
	if err != nil {
		fmt.Printf("Failed to validate update on user, error: %v", err)
	}

	return &existUser, nil

}

func UserGetReview(userID string) (*models.UserReviews, error) {
	var PRSlice []models.PullRequestShort

	err := DB.Select(
		&PRSlice,
		`SELECT pull_request_id, pull_request_name, author_id, status 
		FROM pull_requests 
		WHERE $1 = ANY(assigned_reviewers)`, userID)
	if err != nil {
		fmt.Printf("Failed search for user in database. Error: %v", err)
		return nil, err
	}

	return &models.UserReviews{
		UserID:       userID,
		PullRequests: PRSlice,
	}, nil
}

var assigneeQuery string = `
WITH filtered_users AS (
  SELECT user_id
  FROM users
  INNER JOIN teams ON users.team_name = $1
  WHERE NOT EXISTS (
    SELECT 1
    FROM pull_requests, unnest(assigned_reviewers) AS reviewer
    WHERE reviewer = users.user_id AND pull_requests.status <> 'MERGED'
  )
  AND users.is_active = true
	AND users.user_id <> $2
)
SELECT array_agg(user_id) AS user_id, (SELECT COUNT(*) FROM filtered_users) AS total_count
FROM filtered_users
`

type potentialAssignees struct {
	UserIDs    pq.StringArray `db:"user_id"`
	TotalCount int            `db:"total_count"`
}

func CreatePR(newPr *dto.CreatePR) (*models.PullRequest, error) {
	var numOfAuthor int
	err := DB.Get(&numOfAuthor, "SELECT COUNT(*) FROM users WHERE user_id = $1", newPr.AuthorID)
	if err != nil {
		fmt.Printf("Failed to fetch authors from database, error: %v", err)
		return nil, err
	}
	if numOfAuthor == 0 {
		return nil, errors.New("NOT_FOUND")
	}

	var numOfPr int
	err = DB.Get(&numOfPr, "SELECT COUNT(*) FROM pull_requests WHERE pull_request_id = $1", newPr.PRID)
	if err != nil {
		fmt.Printf("Failed to fetch pull requests from database, error: %v", err)
		return nil, err
	}
	if numOfPr > 0 {
		return nil, errors.New("PR_EXISTS")
	}

	var allAssignees potentialAssignees
	var authorTeam string

	err = DB.Get(&authorTeam, "SELECT team_name FROM users WHERE user_id = $1", newPr.AuthorID)
	if err != nil {
		fmt.Printf("Can't fetch author's team, error: %v", err)
		return nil, err
	}

	err = DB.Get(&allAssignees, assigneeQuery, authorTeam, newPr.AuthorID)

	if err != nil {
		fmt.Printf("Can't fetch potential assignees, error: %v", err)
		return nil, err
	}

	var newAssignees []string

	switch allAssignees.TotalCount {
	case 0:
		break
	case 1:
		newAssignees = append(newAssignees, allAssignees.UserIDs[0])
	default:
		first := rand.Intn(len(allAssignees.UserIDs))
		second := rand.Intn(len(allAssignees.UserIDs))
		for second == first {
			second = rand.Intn(len(allAssignees.UserIDs))
		}
		newAssignees = append(newAssignees, allAssignees.UserIDs[first])
		newAssignees = append(newAssignees, allAssignees.UserIDs[second])
	}

	var resultedPr models.PullRequest

	err = DB.Get(
		&resultedPr,
		`INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at)
	 	 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING pull_request_id, pull_request_name, author_id, status, assigned_reviewers`,
		newPr.PRID,
		newPr.PRName,
		newPr.AuthorID,
		"OPEN",
		pq.StringArray(newAssignees),
		time.Now(),
	)

	if err != nil {
		fmt.Printf("Error inserting new pull request, %v\n", err)
		return nil, err
	} else {
		fmt.Printf("New pull request %s has been registered\n", newPr.PRName)
	}

	return &resultedPr, nil
}

func MergePR(requestId *models.RequestPullRequestMerge) (*models.PullRequest, error) {
	PRID := requestId.PullRequestId
	var rowCount int
	err := DB.Get(&rowCount, "SELECT COUNT(*) FROM pull_requests WHERE pull_request_id = $1", PRID)
	if err != nil {
		fmt.Printf("Failed to fetch pull requests from database, error: %v", err)
		return nil, err
	}
	if rowCount == 0 {
		return nil, errors.New("NOT_FOUND")
	}

	var mergedPR models.PullRequest

	mergeTime := time.Now()
	err = DB.Get(
		&mergedPR,
		`UPDATE pull_requests SET status = 'MERGED', merged_at = $1
	 	 WHERE pull_request_id = $2
		 RETURNING pull_request_id, pull_request_name, author_id, status, assigned_reviewers, merged_at`,
		mergeTime,
		PRID)
	if err != nil {
		fmt.Printf("Failed to update pull request status, error: %v", err)
		return nil, err
	}
	return &mergedPR, nil
}

var reassignQuery string = `
WITH current_pr AS (
  SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at
  FROM pull_requests
  WHERE pull_request_id = $2
),
random_reviewer AS (
  SELECT user_id
  FROM users, current_pr
  WHERE team_name = (SELECT team_name FROM users WHERE user_id = $1)
    AND user_id <> $1
    AND user_id <> ALL(current_pr.assigned_reviewers)
		AND is_active = true
  ORDER BY RANDOM()
  LIMIT 1
),
updated_pr AS (
  SELECT
    pull_request_id,
    pull_request_name,
    author_id,
    status,
    array_replace(
      assigned_reviewers,
      $1,
      (SELECT user_id FROM random_reviewer)
    ) AS assigned_reviewers,
    created_at,
    merged_at,
    (SELECT user_id FROM random_reviewer) AS replaced_by
  FROM current_pr
)
UPDATE pull_requests pr
SET assigned_reviewers = u.assigned_reviewers
FROM updated_pr u
WHERE pr.pull_request_id = u.pull_request_id
RETURNING u.pull_request_id, u.pull_request_name, u.author_id, u.status, u.assigned_reviewers, 
          u.created_at, u.merged_at, u.replaced_by


`

func PrReassign(requestData *models.RequestPrReassign) (*models.ResponsePrReassign, error) {

	var prCount int
	err := DB.Get(&prCount, "SELECT COUNT(*) FROM pull_requests WHERE pull_request_id = $1", requestData.PullRequestID)
	if err != nil {
		return nil, errors.New("NOT_FOUND")
	}

	var userCount int
	err = DB.Get(&userCount, "SELECT COUNT(*) FROM users WHERE user_id = $1", requestData.OldReviewerID)
	if err != nil {
		return nil, errors.New("NOT_FOUND")
	}

	var prStatus string
	err = DB.Get(&prStatus, "SELECT status FROM pull_requests WHERE pull_request_id = $1", requestData.PullRequestID)
	if err != nil {
		fmt.Printf("Failed to fetch pull request status, error: %v", err)
		return nil, err
	}
	if prStatus == "MERGED" {
		return nil, errors.New("PR_MERGED")
	}

	var isAssigned bool
	err = DB.Get(
		&isAssigned,
		`SELECT $1 = ANY(assigned_reviewers) 
		 FROM pull_requests
		 WHERE pull_request_id = $2`,
		requestData.OldReviewerID,
		requestData.PullRequestID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("NOT_FOUND")
		}
		return nil, err
	}
	if !isAssigned {
		return nil, errors.New("NOT_ASSIGNED")
	}

	var candidateCount int
	err = DB.Get(
		&candidateCount,
		`WITH current_pr AS (
  	 SELECT assigned_reviewers
  	 FROM pull_requests
  	 WHERE pull_request_id = $2
		)
		 SELECT COUNT(*)
	   FROM users, current_pr
		 WHERE team_name = (
		 	SELECT team_name 
			FROM users 
			WHERE user_id = $1)
  	 	AND user_id <> $1
  	 	AND user_id <> ALL(current_pr.assigned_reviewers)
			AND is_active = true`,
		requestData.OldReviewerID,
		requestData.PullRequestID,
	)
	if err != nil {
		fmt.Printf("Failed to fetch available candidates, error: %v", err)
		return nil, err
	}
	if candidateCount == 0 {
		return nil, errors.New("NO_CANDIDATE")
	}

	var mPR models.MiddlePrReassign
	err = DB.Get(&mPR, reassignQuery, requestData.OldReviewerID, requestData.PullRequestID)
	if err != nil {
		fmt.Printf("Failed to reassign reviewer, error: %v", err)
		return nil, err
	}

	reassignedPr := models.ResponsePrReassign{
		ResponsePullRequest: models.ResponsePullRequest{
			PR: models.PullRequest{
				UserPullRequest: models.UserPullRequest{
					PullRequestID:     mPR.PullRequestID,
					PullRequestName:   mPR.PullRequestName,
					AuthorID:          mPR.AuthorID,
					Status:            mPR.Status,
					AssignedReviewers: mPR.AssignedReviewers,
				},
			},
		},
		ReplacedBy: mPR.ReplacedBy,
	}

	return &reassignedPr, nil
}
