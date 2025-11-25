package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"prmanagement/api/dto"
	"prmanagement/db"
	"prmanagement/db/models"
)

func AddTeam(w http.ResponseWriter, r *http.Request) {
	var body models.Team

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleErrors(errors.New("INVALID_REQUEST"), w)
		return
	}

	if err = db.AddTeam(&body); err != nil {
		HandleErrors(errors.New("TEAM_EXISTS"), w)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")

	w.Header().Set("Content-Type", "application/json")

	targetTeam, err := db.GetTeam(teamName)
	if err != nil {
		HandleErrors(err, w)
		return
	}

	jsonData, err := json.Marshal(targetTeam)
	if err != nil {
		HandleErrors(err, w)
		return
	}
	w.Write(jsonData)
}

func UserSetIsActive(w http.ResponseWriter, r *http.Request) {
	var body dto.UserSetIsActive

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleErrors(errors.New("INVALID_REQUEST"), w)
		return
	}

	var resultedUser *models.User
	resultedUser, err = db.UserSetIsActive(&body)
	if err != nil {
		HandleErrors(err, w)
		return
	}

	jsonData, err := json.Marshal(resultedUser)
	if err != nil {
		HandleErrors(err, w)
		return
	}

	w.Write(jsonData)
}

func UserGetPrs(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")

	w.Header().Set("Content-Type", "application/json")

	userReviews, err := db.UserGetReview(userId)
	if err != nil {
		HandleErrors(err, w)
		return
	}

	jsonData, err := json.Marshal(userReviews)
	if err != nil {
		HandleErrors(err, w)
		return
	}

	w.Write(jsonData)
}

func PrCreate(w http.ResponseWriter, r *http.Request) {
	var body dto.CreatePR

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleErrors(errors.New("INVALID_REQUEST"), w)
		return
	}

	newPr, err := db.CreatePR(&body)

	if err != nil {
		HandleErrors(err, w)
	}

	jsonData, err := json.Marshal(models.ResponsePullRequest{*newPr})
	if err != nil {
		HandleErrors(err, w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func PrMerge(w http.ResponseWriter, r *http.Request) {
	var body models.RequestPullRequestMerge

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleErrors(errors.New("INVALID_REQUEST"), w)
		return
	}

	updPr, err := db.MergePR(&body)

	if err != nil {
		HandleErrors(err, w)
	}

	jsonData, err := json.Marshal(updPr)
	if err != nil {
		HandleErrors(err, w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func PrReassign(w http.ResponseWriter, r *http.Request) {
	var body models.RequestPrReassign

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleErrors(errors.New("INVALID_REQUEST"), w)
		return
	}

	reassignedPr, err := db.PrReassign(&body)
	if err != nil {
		HandleErrors(err, w)
		return
	}

	jsonData, err := json.Marshal(*reassignedPr)
	if err != nil {
		HandleErrors(err, w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}
