package dashboard

import (
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/nradojcic/sentinel/internal/store"
)

func TestMain(m *testing.M) {
	// discard all log output during tests to keep the output clean
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	// run all tests
	os.Exit(m.Run())
}

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
	// ensure that a nil template results in an internal server error
	dashboardTemplate = nil // force an error condition

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

func TestNewServer(t *testing.T) {
	myStore := store.NewMonitorStore()
	templatePath := "../../web/index.html"
	port := "8081"

	t.Run("Success", func(t *testing.T) {
		server, err := NewServer(port, templatePath, myStore)

		if err != nil {
			t.Errorf("Wanted no error, but got %v", err)
		}
		if server == nil {
			t.Fatal("Wanted server to be non-nil")
		}
		if server.Addr != ":"+port {
			t.Errorf("Wanted server address to be ':%s', but got '%s'", port, server.Addr)
		}
		if server.Handler == nil {
			t.Error("Wanted server handler to be non-nil")
		}
	})

	t.Run("TemplateNotFound", func(t *testing.T) {
		server, err := NewServer(port, "non-existent-template.html", myStore)

		if err == nil {
			t.Error("Wanted an error for non-existent template, but got nil")
		}
		if server != nil {
			t.Error("Wanted server to be nil on template error")
		}
	})
}
