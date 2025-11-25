package api

import "net/http"

func RegisterAPI() error {
	http.HandleFunc("POST /team/add", AddTeam)
	http.HandleFunc("GET /team/get", GetTeam)

	http.HandleFunc("POST /users/setIsActive", UserSetIsActive)
	
	http.HandleFunc("POST /pullRequest/create", PrCreate)
	http.HandleFunc("POST /pullRequest/merge", PrMerge)

	return nil
}
