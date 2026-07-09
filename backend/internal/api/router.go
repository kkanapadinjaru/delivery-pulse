package api

import (
	"encoding/json"
	"net/http"

	"github.com/delivery-pulse/backend/internal/ado"
	"github.com/delivery-pulse/backend/internal/settings"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter creates the HTTP router with all API routes.
func NewRouter(client *ado.Client, store *settings.Store) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5180", "http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	h := &handler{client: client, settings: store}

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", h.health)
		r.Get("/developers", h.getDevelopers)
		r.Get("/report", h.getReport)
		r.Get("/workitems", h.getWorkItems)
		r.Get("/settings", h.getSettings)
		r.Put("/settings", h.updateSettings)
	})

	return r
}

type handler struct {
	client   *ado.Client
	settings *settings.Store
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) getDevelopers(w http.ResponseWriter, r *http.Request) {
	developers, err := h.client.GetDevelopers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"developers": developers,
	})
}

func (h *handler) getReport(w http.ResponseWriter, r *http.Request) {
	developer := r.URL.Query().Get("developer")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if developer == "" || from == "" || to == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "developer, from, and to query parameters are required",
		})
		return
	}

	report, err := h.client.GetDeveloperReport(developer, from, to)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func (h *handler) getWorkItems(w http.ResponseWriter, r *http.Request) {
	developer := r.URL.Query().Get("developer")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if developer == "" || from == "" || to == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "developer, from, and to query parameters are required",
		})
		return
	}

	items, err := h.client.GetWorkItems(developer, from, to)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"workItems": items,
		"count":     len(items),
	})
}

func (h *handler) getSettings(w http.ResponseWriter, r *http.Request) {
	s := h.settings.Get()
	writeJSON(w, http.StatusOK, s)
}

func (h *handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var s settings.Settings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	if err := h.settings.Update(s); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to save settings: " + err.Error(),
		})
		return
	}

	// Apply teams change to the ADO client immediately
	h.client.SetTeams(s.Teams)

	// Apply work item types change
	h.client.SetWorkItemTypes(s.WorkItemTypes)

	writeJSON(w, http.StatusOK, h.settings.Get())
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
