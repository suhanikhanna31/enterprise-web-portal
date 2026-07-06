package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Campaign struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Budget    float64 `json:"budget"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at,omitempty"`
}

type PerformanceSummary struct {
	Impressions int     `json:"impressions"`
	Clicks      int     `json:"clicks"`
	CTR         float64 `json:"ctr"`
	Revenue     float64 `json:"revenue"`
}

// GET /api/campaigns
func GetCampaigns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rows, err := DB.Query("SELECT id, name, budget, status FROM campaigns ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var campaigns []Campaign = []Campaign{}
	for rows.Next() {
		var c Campaign
		if err := rows.Scan(&c.ID, &c.Name, &c.Budget, &c.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		campaigns = append(campaigns, c)
	}

	json.NewEncoder(w).Encode(campaigns)
}

// POST /api/campaigns
func CreateCampaign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var c Campaign
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid input payload", http.StatusBadRequest)
		return
	}

	err := DB.QueryRow(
		"INSERT INTO campaigns (name, budget, status) VALUES ($1, $2, $3) RETURNING id",
		c.Name, c.Budget, "active",
	).Scan(&c.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Status = "active"
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// GET /api/analytics
func GetAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Simulating aggregate analytical queries typical of BigQuery/ClickHouse pipelines
	query := `
		SELECT 
			COUNT(CASE WHEN event_type = 'impression' THEN 1 END) as impressions,
			COUNT(CASE WHEN event_type = 'click' THEN 1 END) as clicks,
			COALESCE(SUM(revenue), 0) as total_revenue
		FROM ad_events`

	var summary PerformanceSummary
	err := DB.QueryRow(query).Scan(&summary.Impressions, &summary.Clicks, &summary.Revenue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if summary.Impressions > 0 {
		summary.CTR = (float64(summary.Clicks) / float64(summary.Impressions)) * 100
	}

	json.NewEncoder(w).Encode(summary)
}

// POST /api/track (Simulating edge real-time tracking pixels)
func TrackEvent(w http.ResponseWriter, r *http.Request) {
	campaignIDStr := r.URL.Query().Get("campaign_id")
	eventType := r.URL.Query().Get("type") // 'impression' or 'click'
	
	campaignID, err := strconv.Atoi(campaignIDStr)
	if err != nil || (eventType != "impression" && eventType != "click") {
		http.Error(w, "Missing or invalid parameters", http.StatusBadRequest)
		return
	}

	var revenue float64 = 0.002 // Default micro-revenue for ad impression
	if eventType == "click" {
		revenue = 0.45 // Higher payout click event
	}

	_, err = DB.Exec("INSERT INTO ad_events (campaign_id, event_type, revenue) VALUES ($1, $2, $3)", campaignID, eventType, revenue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}
