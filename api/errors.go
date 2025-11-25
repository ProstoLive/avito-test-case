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
