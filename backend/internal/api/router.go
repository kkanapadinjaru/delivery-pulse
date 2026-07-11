package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/delivery-pulse/backend/internal/ado"
	"github.com/delivery-pulse/backend/internal/settings"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter creates the HTTP router with all API routes.
func NewRouter(client *ado.Client, store *settings.Store, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(slogRequestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5180", "http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	h := &handler{client: client, settings: store, logger: logger.With("component", "api")}

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", h.health)
		r.Get("/developers", h.getDevelopers)
		r.Get("/report", h.getReport)
		r.Get("/workitems", h.getWorkItems)
		r.Get("/areapaths", h.getAreaPaths)
		r.Get("/settings", h.getSettings)
		r.Put("/settings", h.updateSettings)
	})

	return r
}

// slogRequestLogger returns a middleware that logs each request with structured fields.
func slogRequestLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	reqLogger := logger.With("component", "http")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			status := ww.Status()

			level := slog.LevelInfo
			if status >= 500 {
				level = slog.LevelError
			} else if status >= 400 {
				level = slog.LevelWarn
			}

			reqLogger.Log(r.Context(), level, "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"bytes", ww.BytesWritten(),
				"remote", r.RemoteAddr,
			)
		})
	}
}

type handler struct {
	client   *ado.Client
	settings *settings.Store
	logger   *slog.Logger
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) getDevelopers(w http.ResponseWriter, r *http.Request) {
	developers, err := h.client.GetDevelopers()
	if err != nil {
		h.logger.Error("failed to fetch developers", "error", err)
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
		h.logger.Error("failed to generate report",
			"developer", developer,
			"from", from,
			"to", to,
			"error", err,
		)
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
		h.logger.Error("failed to fetch work items",
			"developer", developer,
			"from", from,
			"to", to,
			"error", err,
		)
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

func (h *handler) getAreaPaths(w http.ResponseWriter, r *http.Request) {
	paths, err := h.client.GetAreaPaths()
	if err != nil {
		h.logger.Error("failed to fetch area paths", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"areaPaths": paths,
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
		h.logger.Error("failed to save settings", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to save settings: " + err.Error(),
		})
		return
	}

	// Apply teams change to the ADO client immediately
	h.client.SetTeams(s.Teams)

	// Apply work item types change
	h.client.SetWorkItemTypes(s.WorkItemTypes)

	// Apply area paths and activities
	h.client.SetAreaPaths(s.AreaPaths)
	h.client.SetActivities(s.Activities)

	h.logger.Info("settings updated",
		"teams", s.Teams,
		"workItemTypes", s.WorkItemTypes,
		"areaPaths", s.AreaPaths,
		"activities", s.Activities,
	)

	writeJSON(w, http.StatusOK, h.settings.Get())
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
