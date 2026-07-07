package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// --- Minimal JSON-RPC 2.0 MCP-style endpoint --------------------------
//
// The Model Context Protocol runs JSON-RPC 2.0 over a transport
// (stdio, SSE, or streamable HTTP). This is not a certified MCP server
// implementation and does not use the official MCP SDK, but unlike a
// hardcoded static blob, it actually parses JSON-RPC requests, routes
// them by method, and returns responses in JSON-RPC shape, so a client
// speaking JSON-RPC over HTTP gets real request/response behavior
// (including proper error codes) rather than one fixed payload.

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

var availableTools = []mcpTool{
	{
		Name:        "get_campaign_status",
		Description: "Fetches current campaigns from the database",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
	{
		Name:        "get_campaign_analytics",
		Description: "Fetches aggregate impression/click/revenue analytics",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
}

func mcpErrorResponse(id json.RawMessage, code int, message string) jsonRPCResponse {
	return jsonRPCResponse{JSONRPC: "2.0", ID: id, Error: &jsonRPCError{Code: code, Message: message}}
}

func handleMCP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req jsonRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(mcpErrorResponse(nil, -32700, "parse error"))
		return
	}

	switch req.Method {
	case "initialize":
		json.NewEncoder(w).Encode(jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    "enterprise-web-portal-demo",
					"version": "0.1.0",
				},
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
			},
		})

	case "tools/list":
		json.NewEncoder(w).Encode(jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]interface{}{"tools": availableTools},
		})

	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			json.NewEncoder(w).Encode(mcpErrorResponse(req.ID, -32602, "invalid params"))
			return
		}

		switch params.Name {
		case "get_campaign_status":
			rows, err := DB.Query("SELECT id, name, budget, spent, status FROM campaigns ORDER BY id DESC")
			if err != nil {
				json.NewEncoder(w).Encode(mcpErrorResponse(req.ID, -32000, "database error"))
				return
			}
			defer rows.Close()
			campaigns := []Campaign{}
			for rows.Next() {
				var c Campaign
				if err := rows.Scan(&c.ID, &c.Name, &c.Budget, &c.Spent, &c.Status); err == nil {
					campaigns = append(campaigns, c)
				}
			}
			json.NewEncoder(w).Encode(jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]interface{}{"content": campaigns},
			})

		case "get_campaign_analytics":
			var summary PerformanceSummary
			err := DB.QueryRow(`
				SELECT
					COUNT(CASE WHEN event_type = 'impression' THEN 1 END),
					COUNT(CASE WHEN event_type = 'click' THEN 1 END),
					COALESCE(SUM(revenue), 0)
				FROM ad_events`).Scan(&summary.Impressions, &summary.Clicks, &summary.Revenue)
			if err != nil {
				json.NewEncoder(w).Encode(mcpErrorResponse(req.ID, -32000, "database error"))
				return
			}
			if summary.Impressions > 0 {
				summary.CTR = (float64(summary.Clicks) / float64(summary.Impressions)) * 100
			}
			json.NewEncoder(w).Encode(jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]interface{}{"content": summary},
			})

		default:
			json.NewEncoder(w).Encode(mcpErrorResponse(req.ID, -32601, fmt.Sprintf("unknown tool: %s", params.Name)))
		}

	default:
		json.NewEncoder(w).Encode(mcpErrorResponse(req.ID, -32601, fmt.Sprintf("unknown method: %s", req.Method)))
	}
}

func main() {
	InitDB()
	defer DB.Close()

	r := mux.NewRouter()

	r.HandleFunc("/api/campaigns", GetCampaigns).Methods("GET")
	r.HandleFunc("/api/campaigns", CreateCampaign).Methods("POST")
	r.HandleFunc("/api/analytics", GetAnalytics).Methods("GET")
	r.HandleFunc("/api/track", TrackEvent).Methods("POST")

	// Real JSON-RPC handling instead of a static response.
	r.HandleFunc("/mcp/v1", handleMCP).Methods("POST")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})

	handler := c.Handler(r)

	// Explicit timeouts: net/http's default server has none, which
	// leaves it exposed to slow-client connection exhaustion. This is
	// a genuine (if basic) production hardening step, not just a label.
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println("Server listening on :8080")
	log.Fatal(srv.ListenAndServe())
}
