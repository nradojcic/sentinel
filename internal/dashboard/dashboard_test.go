package dashboard

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nradojcic/sentinel/internal/store"
)

func TestDashboardHandler(t *testing.T) {
	// 1. Initialize a MonitorStore with mock data
	myStore := store.NewMonitorStore()
	myStore.Update("agent-1", 55.5, 78.9)
	myStore.Update("agent-2", 12.3, 45.6)

	// 2. Load the HTML template
	templatePath := "../../web/index.html"
	var err error
	dashboardTemplate, err = template.ParseFiles(templatePath)
	if err != nil {
		t.Fatalf("Failed to parse dashboard template: %v", err)
	}

	// 3. Create a request and response recorder
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// 4. Call the dashboardHandler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dashboardHandler(myStore, w, r)
	})
	handler.ServeHTTP(rr, req)

	// 5. Verify the results
	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check if the response body contains the agent data
	body := rr.Body.String()
	if !strings.Contains(body, "agent-1") {
		t.Errorf("handler response body does not contain 'agent-1'")
	}
	if !strings.Contains(body, "55.5") {
		t.Errorf("handler response body does not contain '55.5'")
	}
	if !strings.Contains(body, "78.9") {
		t.Errorf("handler response body does not contain '78.9'")
	}

	if !strings.Contains(body, "agent-2") {
		t.Errorf("handler response body does not contain 'agent-2'")
	}
	if !strings.Contains(body, "12.3") {
		t.Errorf("handler response body does not contain '12.3'")
	}
	if !strings.Contains(body, "45.6") {
		t.Errorf("handler response body does not contain '45.6'")
	}

	// Check for a piece of static text from the template
	if !strings.Contains(body, "Sentinel Dashboard") {
		t.Errorf("handler response body does not contain 'Sentinel Dashboard'")
	}
}

func TestDashboardHandler_TemplateError(t *testing.T) {
	// Ensure that a nil template results in an internal server error
	dashboardTemplate = nil // Force an error condition

	myStore := store.NewMonitorStore()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dashboardHandler(myStore, w, r)
	})
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code for template error: got %v want %v",
			status, http.StatusInternalServerError)
	}
}
