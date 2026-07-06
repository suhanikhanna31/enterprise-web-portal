package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	InitDB()
	defer DB.Close()

	r := mux.NewRouter()

	// REST API Routes
	r.HandleFunc("/api/campaigns", GetCampaigns).Methods("GET")
	r.HandleFunc("/api/campaigns", CreateCampaign).Methods("POST")
	r.HandleFunc("/api/analytics", GetAnalytics).Methods("GET")
	r.HandleFunc("/api/track", TrackEvent).Methods("POST")

	// Optional Demonstration of MCP Protocol concept endpoint standard
	r.HandleFunc("/mcp/v1/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tools": [
				{"name": "get_campaign_status", "description": "Fetches current live system ad campaigns"},
				{"name": "optimize_budgets", "description": "Triggers automated real-time bidding parameter changes"}
			]
		}`))
	}).Methods("GET")

	// Configure CORS for local development with React
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})

	handler := c.Handler(r)

	fmt.Println("AdTech Engine server running efficiently on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
