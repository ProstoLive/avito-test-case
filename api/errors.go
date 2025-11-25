package api

import (
	"encoding/json"
	"net/http"
)

type RoutesError struct {
	Code    string
	Message string
}

func JsonableError(w http.ResponseWriter, re RoutesError) error {
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"error": re,
	})
}

func HandleErrors(e error, w http.ResponseWriter) error {
	switch e.Error() {
	case "NOT_FOUND":
		w.WriteHeader(http.StatusNotFound)
		return JsonableError(w, RoutesError{
			Code: "NOT_FOUND",
			Message: "resource not found",
		})
	case "PR_MERGED":
		w.WriteHeader(http.StatusConflict)
		return JsonableError(w, RoutesError{
			Code: "PR_MERGED",
			Message: "cannot reassign on merged PR",
		})
	case "NOT_ASSIGNED":
		w.WriteHeader(http.StatusConflict)
		return JsonableError(w, RoutesError{
			Code: "NOT_ASSIGNED",
			Message: "reviewer is not assigned to this PR",
		})
	case "NO_CANDIDATE":
		w.WriteHeader(http.StatusConflict)
		return JsonableError(w, RoutesError{
			Code: "NO_CANDIDATE",
			Message: "no active replacement candidate in team",
		})
	case "PR_EXISTS":
		w.WriteHeader(http.StatusConflict)
		return JsonableError(w, RoutesError{
			Code: "PR_EXISTS",
			Message: "PR id already exists",
		})
	case "TEAM_EXISTS":
		w.WriteHeader(http.StatusBadRequest)
		return JsonableError(w, RoutesError{
			Code: "TEAM_EXISTS",
			Message: "team_name already exists",
		})
	case "INVALID_REQUEST":
		w.WriteHeader(http.StatusBadRequest)
		return JsonableError(w, RoutesError{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		})
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return JsonableError(w, RoutesError{
			Code:    "INTERNAL_SERVER",
			Message: "Internal server error",
		})
	}

}
