package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB opens a connection pool to Postgres. This is a single-instance
// relational database, not a distributed analytics engine. If you need
// real high-cardinality event analytics at scale, this table will need
// to be replaced by (or fed into) a purpose-built system such as
// ClickHouse, BigQuery, or a time-series store — plain Postgres COUNT/SUM
// queries here will not scale past a modest number of rows without
// partitioning, rollups, or a columnar store.
func InitDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "user=postgres password=postgres dbname=adtech_db sslmode=disable"
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	// Real connection-pool tuning. net/http's goroutine-per-request model
	// gives you concurrent request handling for free, but without limits
	// here a burst of traffic can exhaust Postgres' max_connections.
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	if err = DB.Ping(); err != nil {
		log.Fatal("Database unreachable: ", err)
	}

	fmt.Println("Connected to PostgreSQL.")
	createTables()
}

func createTables() {
	campaignTable := `
	CREATE TABLE IF NOT EXISTS campaigns (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		budget NUMERIC(10, 2) NOT NULL CHECK (budget > 0),
		spent NUMERIC(10, 4) NOT NULL DEFAULT 0,
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Plain relational events table. Fine for a demo or low volume;
	// add the indexes below so at least the two queries the app
	// actually runs (per-campaign lookups, time-ordered scans) aren't
	// full table scans as row counts grow.
	eventsTable := `
	CREATE TABLE IF NOT EXISTS ad_events (
		id SERIAL PRIMARY KEY,
		campaign_id INT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
		event_type VARCHAR(20) NOT NULL CHECK (event_type IN ('impression', 'click')),
		revenue NUMERIC(10, 4) NOT NULL DEFAULT 0.0000,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_ad_events_campaign_id ON ad_events (campaign_id);`,
		`CREATE INDEX IF NOT EXISTS idx_ad_events_timestamp ON ad_events (timestamp);`,
	}

	if _, err := DB.Exec(campaignTable); err != nil {
		log.Fatal("Error creating campaigns table: ", err)
	}
	if _, err := DB.Exec(eventsTable); err != nil {
		log.Fatal("Error creating ad_events table: ", err)
	}
	for _, idx := range indexes {
		if _, err := DB.Exec(idx); err != nil {
			log.Fatal("Error creating index: ", err)
		}
	}
}
