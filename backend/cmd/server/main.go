package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/delivery-pulse/backend/internal/api"
	"github.com/delivery-pulse/backend/internal/ado"
	"github.com/delivery-pulse/backend/internal/settings"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load .env file if it exists (local development)
	// Try multiple paths to support running from backend/ or project root
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	orgURL := os.Getenv("ADO_ORG_URL")
	project := os.Getenv("ADO_PROJECT")
	port := os.Getenv("SERVER_PORT")
	teamsEnv := os.Getenv("ADO_TEAMS")
	developersEnv := os.Getenv("ADO_DEVELOPERS")
	clientID := os.Getenv("ADO_CLIENT_ID")
	clientSecret := os.Getenv("ADO_CLIENT_SECRET")
	tenantID := os.Getenv("ADO_TENANT_ID")

	if orgURL == "" || project == "" {
		logger.Error("missing required environment variables", "required", "ADO_ORG_URL, ADO_PROJECT")
		os.Exit(1)
	}
	if clientID == "" || clientSecret == "" || tenantID == "" {
		logger.Error("missing required environment variables", "required", "ADO_CLIENT_ID, ADO_CLIENT_SECRET, ADO_TENANT_ID")
		os.Exit(1)
	}

	if port == "" {
		port = "8080"
	}

	// Parse comma-separated team names from env (used as initial default)
	var envTeams []string
	if teamsEnv != "" {
		for _, t := range strings.Split(teamsEnv, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				envTeams = append(envTeams, t)
			}
		}
	}

	// Parse comma-separated developer emails from env (used as initial default)
	var envDevelopers []string
	if developersEnv != "" {
		for _, d := range strings.Split(developersEnv, ",") {
			d = strings.TrimSpace(d)
			if d != "" {
				envDevelopers = append(envDevelopers, d)
			}
		}
	}

	// Initialize settings store
	execPath, _ := os.Executable()
	settingsDir := filepath.Dir(execPath)
	settingsFile := filepath.Join(settingsDir, "settings.json")
	store := settings.NewStore(settingsFile)

	// If settings file has no teams but env does, seed from env
	currentSettings := store.Get()
	seeded := false
	if len(currentSettings.Teams) == 0 && len(envTeams) > 0 {
		currentSettings.Teams = envTeams
		seeded = true
	}
	if len(currentSettings.Developers) == 0 && len(envDevelopers) > 0 {
		currentSettings.Developers = envDevelopers
		seeded = true
	}
	if seeded {
		_ = store.Update(currentSettings)
	}

	client := ado.NewClient(orgURL, tenantID, clientID, clientSecret, project, logger)
	client.SetTeams(store.Get().Teams)
	client.SetDevelopers(store.Get().Developers)
	client.SetWorkItemTypes(store.Get().WorkItemTypes)
	client.SetAreaPaths(store.Get().AreaPaths)
	client.SetActivities(store.Get().Activities)
	client.SetPRSizeThresholds(store.Get().PRSizeSmallMax, store.Get().PRSizeMediumMax)

	router := api.NewRouter(client, store, logger)

	addr := fmt.Sprintf(":%s", port)
	logger.Info("server starting",
		"addr", addr,
		"org", orgURL,
		"project", project,
		"auth", "service-principal",
	)
	appliedSettings := store.Get()
	if len(appliedSettings.Teams) > 0 {
		logger.Info("configuration loaded",
			"teams", appliedSettings.Teams,
			"workItemTypes", appliedSettings.WorkItemTypes,
		)
	}

	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
