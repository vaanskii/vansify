package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lpernett/godotenv"
)

var DB *sql.DB

// ConnectToDatabase initializes the database connection
func ConnectToDatabase() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get environment variables
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	// Remove local DB variable declaration
	var err error // Declare err here
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Verify the connection
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	fmt.Println("Successfully connected to MySQL database!")
}


// Automatically making migrations while server runing if users table not exists
func CreateTable() {
	// Create the users table
	query := `CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(100) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		email VARCHAR(100) NOT NULL UNIQUE
	);`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}