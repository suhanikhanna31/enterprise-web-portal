package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func newMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	original := DB
	DB = db
	return mock, func() { DB = original; db.Close() }
}

func TestCreateCampaign_RejectsMissingName(t *testing.T) {
	body := []byte(`{"budget": 100}`)
	req := httptest.NewRequest(http.MethodPost, "/api/campaigns", bytes.NewReader(body))
	w := httptest.NewRecorder()

	CreateCampaign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", w.Code)
	}
}

func TestCreateCampaign_RejectsZeroBudget(t *testing.T) {
	body := []byte(`{"name": "Test Campaign", "budget": 0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/campaigns", bytes.NewReader(body))
	w := httptest.NewRecorder()

	CreateCampaign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for zero budget, got %d", w.Code)
	}
}

func TestCreateCampaign_Success(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "spent"}).AddRow(1, 0.0)
	mock.ExpectQuery("INSERT INTO campaigns").
		WithArgs("Test Campaign", 100.0).
		WillReturnRows(rows)

	body := []byte(`{"name": "Test Campaign", "budget": 100}`)
	req := httptest.NewRequest(http.MethodPost, "/api/campaigns", bytes.NewReader(body))
	w := httptest.NewRecorder()

	CreateCampaign(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body: %s", w.Code, w.Body.String())
	}

	var got Campaign
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.ID != 1 || got.Status != "active" {
		t.Fatalf("unexpected campaign in response: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestTrackEvent_RejectsInvalidType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/track?campaign_id=1&type=bogus", nil)
	w := httptest.NewRecorder()

	TrackEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid event type, got %d", w.Code)
	}
}

func TestTrackEvent_RejectsMissingCampaignID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/track?type=click", nil)
	w := httptest.NewRecorder()

	TrackEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing campaign_id, got %d", w.Code)
	}
}

func TestTrackEvent_CampaignNotFound(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT budget, spent, status FROM campaigns").
		WithArgs(999).
		WillReturnError(errNoRowsForTest())
	mock.ExpectRollback()

	req := httptest.NewRequest(http.MethodPost, "/api/track?campaign_id=999&type=click", nil)
	w := httptest.NewRecorder()

	TrackEvent(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown campaign, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestGetAnalytics_ComputesCTR(t *testing.T) {
	mock, cleanup := newMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"impressions", "clicks", "total_revenue"}).
		AddRow(100, 10, 5.5)
	mock.ExpectQuery("SELECT(.|\n)*FROM ad_events").WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics", nil)
	w := httptest.NewRecorder()

	GetAnalytics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var summary PerformanceSummary
	if err := json.Unmarshal(w.Body.Bytes(), &summary); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if summary.CTR != 10.0 {
		t.Errorf("expected CTR of 10.0, got %v", summary.CTR)
	}
}

// errNoRowsForTest returns sql.ErrNoRows without importing database/sql
// directly into every test that needs it (kept local for readability).
func errNoRowsForTest() error {
	return sql.ErrNoRows
}
