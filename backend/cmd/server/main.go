package main

import (
	"fmt"
	"log"
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
	// Load .env file if it exists (local development)
	// Try multiple paths to support running from backend/ or project root
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	orgURL := os.Getenv("ADO_ORG_URL")
	project := os.Getenv("ADO_PROJECT")
	port := os.Getenv("SERVER_PORT")
	teamsEnv := os.Getenv("ADO_TEAMS")
	clientID := os.Getenv("ADO_CLIENT_ID")
	clientSecret := os.Getenv("ADO_CLIENT_SECRET")
	tenantID := os.Getenv("ADO_TENANT_ID")

	if orgURL == "" || project == "" {
		log.Fatal("ADO_ORG_URL and ADO_PROJECT environment variables are required")
	}
	if clientID == "" || clientSecret == "" || tenantID == "" {
		log.Fatal("ADO_CLIENT_ID, ADO_CLIENT_SECRET, and ADO_TENANT_ID environment variables are required")
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

	// Initialize settings store
	execPath, _ := os.Executable()
	settingsDir := filepath.Dir(execPath)
	settingsFile := filepath.Join(settingsDir, "settings.json")
	store := settings.NewStore(settingsFile)

	// If settings file has no teams but env does, seed from env
	currentSettings := store.Get()
	if len(currentSettings.Teams) == 0 && len(envTeams) > 0 {
		currentSettings.Teams = envTeams
		_ = store.Update(currentSettings)
	}

	client := ado.NewClient(orgURL, tenantID, clientID, clientSecret, project)
	client.SetTeams(store.Get().Teams)
	client.SetWorkItemTypes(store.Get().WorkItemTypes)

	router := api.NewRouter(client, store)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Connected to Azure DevOps: %s / %s (Service Principal)", orgURL, project)
	appliedSettings := store.Get()
	if len(appliedSettings.Teams) > 0 {
		log.Printf("Configured teams: %s", strings.Join(appliedSettings.Teams, ", "))
	}
	log.Printf("Work item types: %s", strings.Join(appliedSettings.WorkItemTypes, ", "))

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
