package api

import "net/http"

func RegisterAPI() {
	http.HandleFunc("POST /team/add", AddTeam)
	http.HandleFunc("GET /team/get", GetTeam)

	http.HandleFunc("POST /users/setIsActive", UserSetIsActive)
	http.HandleFunc("GET /users/getReview", UserGetPrs)

	http.HandleFunc("POST /pullRequest/create", PrCreate)
	http.HandleFunc("POST /pullRequest/merge", PrMerge)
	http.HandleFunc("POST /pullRequest/reassign", PrReassign)
}
