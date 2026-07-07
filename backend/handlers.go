package main

import (
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Campaign struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Budget    float64 `json:"budget"`
	Spent     float64 `json:"spent"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at,omitempty"`
}

type PerformanceSummary struct {
	Impressions int     `json:"impressions"`
	Clicks      int     `json:"clicks"`
	CTR         float64 `json:"ctr"`
	Revenue     float64 `json:"revenue"`
}

type apiError struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiError{Error: msg})
}

// GET /api/campaigns
func GetCampaigns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rows, err := DB.Query("SELECT id, name, budget, spent, status FROM campaigns ORDER BY id DESC")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to query campaigns")
		return
	}
	defer rows.Close()

	campaigns := []Campaign{}
	for rows.Next() {
		var c Campaign
		if err := rows.Scan(&c.ID, &c.Name, &c.Budget, &c.Spent, &c.Status); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to read campaign row")
			return
		}
		campaigns = append(campaigns, c)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error iterating campaigns")
		return
	}

	json.NewEncoder(w).Encode(campaigns)
}

// POST /api/campaigns
func CreateCampaign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var c Campaign
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if c.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}
	if c.Budget <= 0 {
		writeJSONError(w, http.StatusBadRequest, "budget must be greater than zero")
		return
	}

	err := DB.QueryRow(
		"INSERT INTO campaigns (name, budget, status) VALUES ($1, $2, 'active') RETURNING id, spent",
		c.Name, c.Budget,
	).Scan(&c.ID, &c.Spent)

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create campaign")
		return
	}

	c.Status = "active"
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// GET /api/analytics
func GetAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// A plain aggregate query over a Postgres table. This is fine at
	// low-to-moderate row counts; it is not a stand-in for a real
	// analytics pipeline (see the note in database.go).
	query := `
		SELECT
			COUNT(CASE WHEN event_type = 'impression' THEN 1 END) as impressions,
			COUNT(CASE WHEN event_type = 'click' THEN 1 END) as clicks,
			COALESCE(SUM(revenue), 0) as total_revenue
		FROM ad_events`

	var summary PerformanceSummary
	err := DB.QueryRow(query).Scan(&summary.Impressions, &summary.Clicks, &summary.Revenue)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to compute analytics")
		return
	}

	if summary.Impressions > 0 {
		summary.CTR = (float64(summary.Clicks) / float64(summary.Impressions)) * 100
	}

	json.NewEncoder(w).Encode(summary)
}

// simulateBidRevenue produces a randomized revenue figure within a
// plausible CPM/CPC-style range. This is explicitly a simulation for
// demo data generation — it is NOT a real-time bidding engine, and does
// not model auctions, floor prices, or advertiser demand. Labeling it
// as such (rather than as two fixed constants) makes that limitation
// visible in the code instead of implying real pricing logic exists.
func simulateBidRevenue(eventType string) float64 {
	if eventType == "click" {
		// Simulated CPC in a rough $0.20-$0.80 range.
		return 0.20 + rand.Float64()*0.60
	}
	// Simulated per-impression revenue in a rough $0.001-$0.004 range.
	return 0.001 + rand.Float64()*0.003
}

// POST /api/track
func TrackEvent(w http.ResponseWriter, r *http.Request) {
	campaignIDStr := r.URL.Query().Get("campaign_id")
	eventType := r.URL.Query().Get("type")

	campaignID, err := strconv.Atoi(campaignIDStr)
	if err != nil || (eventType != "impression" && eventType != "click") {
		writeJSONError(w, http.StatusBadRequest, "missing or invalid parameters: require numeric campaign_id and type=impression|click")
		return
	}

	revenue := simulateBidRevenue(eventType)

	tx, err := DB.Begin()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Lock the campaign row so concurrent track requests can't both read
	// a stale "spent" value and both push the campaign over budget.
	var budget, spent float64
	var status string
	err = tx.QueryRow(
		"SELECT budget, spent, status FROM campaigns WHERE id = $1 FOR UPDATE",
		campaignID,
	).Scan(&budget, &spent, &status)
	if err == sql.ErrNoRows {
		writeJSONError(w, http.StatusNotFound, "campaign not found")
		return
	} else if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to look up campaign")
		return
	}

	if status != "active" {
		writeJSONError(w, http.StatusConflict, "campaign is not active")
		return
	}

	if spent+revenue > budget {
		// Budget exhausted: stop delivering and mark the campaign paused
		// instead of silently accepting the event.
		if _, err := tx.Exec("UPDATE campaigns SET status = 'paused' WHERE id = $1", campaignID); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to pause campaign")
			return
		}
		if err := tx.Commit(); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to commit")
			return
		}
		writeJSONError(w, http.StatusPaymentRequired, "campaign budget exhausted; campaign paused")
		return
	}

	if _, err := tx.Exec(
		"INSERT INTO ad_events (campaign_id, event_type, revenue) VALUES ($1, $2, $3)",
		campaignID, eventType, revenue,
	); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to record event")
		return
	}

	if _, err := tx.Exec(
		"UPDATE campaigns SET spent = spent + $1 WHERE id = $2",
		revenue, campaignID,
	); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to update campaign spend")
		return
	}

	if err := tx.Commit(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to commit transaction")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"revenue": revenue,
	})
}

// getCampaignID is a small helper used by tests/router wiring if a
// path-based lookup (e.g. GET /api/campaigns/{id}) is added later.
func getCampaignID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["id"])
}
