package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"prmanagement/api/dto"
	"prmanagement/db/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
			return &resultTeam, errors.New("NOT_FOUND")
		} else {
			fmt.Printf("Failed to get team from db, error: %v", err)
			return &resultTeam, err
		}
	}
	return &resultTeam, nil
}

func UserSetIsActive(newStatus *dto.UserSetIsActive) (*models.User, error) {
	var existUser models.User
	err := DB.Get(&existUser, "SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1", newStatus.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &existUser, errors.New("NOT_FOUND")
		} else {
			fmt.Printf("Failed search for user in database. Error: %v", err)
			return &existUser, err
		}
	}

	_, err = DB.Exec("UPDATE users SET is_active = $1 WHERE user_id = $2", newStatus.IsActive, newStatus.UserID)

	if err != nil {
		fmt.Printf("Failed update status in database, error: %v", err)
		return &existUser, err
	}

	err = DB.Get(&existUser, "SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1", newStatus.UserID)
	if err != nil {
		fmt.Printf("Failed to validate update on user, error: %v", err)
	}

	return &existUser, nil

}
