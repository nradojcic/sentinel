package dashboard

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/nradojcic/sentinel/internal/store"
)

var dashboardTemplate *template.Template

type DashboardData struct {
	Nodes       []AgentStats
	CurrentTime time.Time
}

type AgentStats struct {
	ID       string
	CPU      float64
	RAM      float64
	LastSeen time.Time
}

// dashboardHandler renders the HTML dashboard page with the latest stats
func dashboardHandler(myStore *store.MonitorStore, w http.ResponseWriter, r *http.Request) {
	if dashboardTemplate == nil {
		http.Error(w, "Dashboard template not loaded", http.StatusInternalServerError)
		slog.Error("Dashboard template is nil")
		return
	}

	allNodesMap := myStore.GetAll()

	// Convert map to slice for sorting and templating
	var nodes []AgentStats
	for id, ns := range allNodesMap {
		nodes = append(nodes, AgentStats{
			ID:       id,
			CPU:      ns.CPU,
			RAM:      ns.RAM,
			LastSeen: ns.LastSeen,
		})
	}

	// Sort nodes by ID for consistent display
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})

	data := DashboardData{
		Nodes:       nodes,
		CurrentTime: time.Now(),
	}

	err := dashboardTemplate.Execute(w, data)
	if err != nil {
		slog.Error("Error executing dashboard template", "error", err)
		http.Error(w, "Error rendering dashboard", http.StatusInternalServerError)
	}
}

// NewServer sets up and returns an HTTP server for the dashboard, but does not start it
func NewServer(port, templatePath string, myStore *store.MonitorStore) (*http.Server, error) {
	// Load the dashboard template
	var err error
	dashboardTemplate, err = template.ParseFiles(templatePath)
	if err != nil {
		slog.Error("Failed to parse dashboard template", "path", templatePath, "error", err)
		return nil, fmt.Errorf("failed to parse dashboard template: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dashboardHandler(myStore, w, r)
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	return httpServer, nil
}
