package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"prmanagement/api/dto"
	"prmanagement/db"
	"prmanagement/db/models"
)

func AddTeam(w http.ResponseWriter, r *http.Request) {
	var body models.Team

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonableError(w, RoutesError{
			Code: "INVALID_REQUEST",
			Message: "Invalid request body",
		})
    return
	}

	if err = db.AddTeam(&body); err != nil {
		fmt.Println("Found error on team addition, ", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		JsonableError(w, RoutesError{
			Code:    "TEAM_EXISTS",
			Message: "team_name already exists",
		})
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")

	w.Header().Set("Content-Type", "application/json")

	targetTeam, err := db.GetTeam(teamName)
	if err != nil {
		fmt.Println(err)
		if err.Error() == "NOT_FOUND" {
			w.WriteHeader(http.StatusBadRequest)

			JsonableError(w, RoutesError{
				Code: "NOT_FOUND",
				Message: "resource not found",
			})
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			JsonableError(w, RoutesError{
				Code: "INTERNAL_SERVER",
				Message: "Internal server error",
			})
			return
		}
	}
		
	jsonData, err := json.Marshal(targetTeam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonableError(w, RoutesError{
			Code: "INTERNAL_SERVER",
			Message: "Internal server error",
		})
		return
	}
	w.Write(jsonData)
}

func UserSetIsActive(w http.ResponseWriter, r *http.Request) {
	var body dto.UserSetIsActive

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonableError(w, RoutesError{
			Code: "INVALID_REQUEST",
			Message: "Invalid request body",
		})
    return
	}

	var resultedUser *models.User
	resultedUser, err = db.UserSetIsActive(&body)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			w.WriteHeader(http.StatusBadRequest)

			JsonableError(w, RoutesError{
				Code: "NOT_FOUND",
				Message: "resource not found",
			})
			return
		}
	}

	jsonData, err := json.Marshal(resultedUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonableError(w, RoutesError{
			Code: "INTERNAL_SERVER",
			Message: "Internal server error",
		})
		return
	}

	w.Write(jsonData)
}

func UserGetPrs(w http.ResponseWriter, r *http.Request) {}

func PrCreate(w http.ResponseWriter, r *http.Request) {}

func PrMerge(w http.ResponseWriter, r *http.Request) {}

func PrReassign(w http.ResponseWriter, r *http.Request) {}