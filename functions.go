package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

var useDatabase bool

func scanner() bool {
	fmt.Println("Will we start the PostgreSQL? (Yes/No)")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := scanner.Text()
		if response == "No" || response == "no" || response == "n" || response == "0" {
			time.Sleep(1 * time.Second)
			fmt.Println("Database will not be started.")
			time.Sleep(3 * time.Second)
			useDatabase = false
			return false
		}
	}

	dbConnectionString := getDBConnection()
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatalf("Error while connecting to database: %s", err)
	}
	defer db.Close()

	dbClient = NewDBClient(db)

	err = dbClient.CreateTable()
	if err != nil {
		log.Fatalf("Error while creating table: %s", err)
	}
	useDatabase = true
	return true
}

func getDBConnection() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)
}
