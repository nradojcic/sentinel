package dashboard

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/nradojcic/sentinel/internal/store"
	"github.com/spf13/viper"
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

// StartServer sets up and starts the HTTP server for the dashboard
func StartServer(myStore *store.MonitorStore) *http.Server {
	httpDashboardPort := viper.GetString("server.http-port")

	// Load the dashboard template once at startup
	templatePath := filepath.Join("web", "index.html")
	var err error
	dashboardTemplate, err = template.ParseFiles(templatePath)
	if err != nil {
		slog.Error("Failed to parse dashboard template", "path", templatePath, "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dashboardHandler(myStore, w, r)
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpDashboardPort),
		Handler: mux,
	}

	slog.Info("Sentinel HTTP dashboard starting", "port", httpDashboardPort)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to serve HTTP dashboard", "error", err)
			os.Exit(1)
		}
	}()

	return httpServer
}
