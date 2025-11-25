package main

import (
	"fmt"
	"net/http"
	"prmanagement/api"
	"prmanagement/db"
)

func main() {
	fmt.Print("Hello avito!\n")

	if err := db.Connect(); err != nil {
		fmt.Print(err)
	}
	fmt.Println("Successfully connected to database")

	api.RegisterAPI()

	http.ListenAndServe(":8080", nil)
}
