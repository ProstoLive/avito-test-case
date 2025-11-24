package main

import (
	"fmt"
	"prmanagement/db"
)

func main() {
	fmt.Print("Hello avito!")
	
	if err := db.Connect(); err != nil {
		fmt.Print(err)
	}
	fmt.Println("Successfully connected to database")
}
