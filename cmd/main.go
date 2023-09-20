package main

import (
	_ "github.com/mattn/go-sqlite3"
	"log"
	"todolist/internal/apiserver"
)

func main() {
	if err := apiserver.Start(); err != nil {
		log.Fatal(err)
	}

}
