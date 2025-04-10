package db

import (
	"GameWala-Arcade/config"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB
var ConfigData []byte

func Initialize() {
	// connStr := "user=username password=password dbname=mydatabase host=localhost sslmode=disable"
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s",
		config.GetString("user"),
		config.GetString("password"),
		config.GetString("name"),
		config.GetString("host"),
		config.GetString("port"))

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	if err = DB.Ping(); err != nil {
		fmt.Println(dsn)
		panic("Database connection error: " + err.Error())
	}

	fmt.Print(dsn)
	fmt.Print("Postgres DB connection established successfully\n\n")
}
