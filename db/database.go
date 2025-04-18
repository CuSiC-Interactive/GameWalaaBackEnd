package db

import (
	"GameWala-Arcade/config"
	"GameWala-Arcade/utils"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB
var ConfigData []byte

func Initialize() {
	utils.LogInfo("Initializing database connection...")
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
		utils.LogError("Failed to connect to database: %v", err)
		panic("Failed to connect to database: " + err.Error())
	}

	if err = DB.Ping(); err != nil {
		utils.LogError("Database ping failed: %v", err)
		panic("Database connection error: " + err.Error())
	}

	utils.LogInfo("Postgres DB connection established successfully")
}
