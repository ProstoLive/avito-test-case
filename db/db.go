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


var AssigneeQuery string = `
WITH filtered_users AS (
  SELECT user_id
  FROM users
  INNER JOIN teams ON users.team_name = $1
  WHERE NOT EXISTS (
    SELECT 1
    FROM pull_requests, unnest(assigned_reviewers) AS reviewer
    WHERE reviewer = users.user_id
  )
  AND users.is_active = true
	AND users.user_id <> $2
)
SELECT array_agg(user_id) AS user_id, (SELECT COUNT(*) FROM filtered_users) AS total_count
FROM filtered_users
`

type potentialAssignees struct {
	UserIDs pq.StringArray `db:"user_id"`
	TotalCount int `db:"total_count"`
}

func CreatePR(newPr *dto.CreatePR) (*models.UserPullRequest, error) {
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

	err = DB.Get(&allAssignees, AssigneeQuery, authorTeam, newPr.AuthorID)

	if err != nil {
		fmt.Printf("Can't fetch potential assignees, error: %v", err)
		return nil, err
	}

	var newAssignees []string

	switch allAssignees.TotalCount{ 
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

	var resultedPr models.UserPullRequest

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
