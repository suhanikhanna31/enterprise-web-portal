package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	// Fallback to local dev credentials if env variables aren't set
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "user=postgres password=postgres dbname=adtech_db sslmode=disable"
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Database unreachable: ", err)
	}

	fmt.Println("Successfully connected to PostgreSQL database!")
	createTables()
}

func createTables() {
	// Campaign Table (Relational)
	campaignTable := `
	CREATE TABLE IF NOT EXISTS campaigns (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		budget NUMERIC(10, 2) NOT NULL,
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Analytics/Impressions Table (Simulating ClickHouse/BigData ingestion structure)
	eventsTable := `
	CREATE TABLE IF NOT EXISTS ad_events (
		id SERIAL PRIMARY KEY,
		campaign_id INT REFERENCES campaigns(id) ON DELETE CASCADE,
		event_type VARCHAR(20) NOT NULL, -- 'impression' or 'click'
		revenue NUMERIC(10, 4) DEFAULT 0.0000,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(campaignTable)
	if err != nil {
		log.Fatal("Error creating campaigns table: ", err)
	}

	_, err = DB.Exec(eventsTable)
	if err != nil {
		log.Fatal("Error creating ad_events table: ", err)
	}
}
