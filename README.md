# Delivery Pulse - Code Contribution & Quality Metrics

A web application to track developer code contribution, velocity, and quality metrics by analyzing Azure DevOps work items and pull requests.

## Purpose

Teams use Azure DevOps to track work items (bugs, tasks, user stories, etc.) and code through pull requests. This app provides visibility into:

- **Developer throughput**: How many work items a developer worked on in a given time period (even if they are no longer the current assignee).
- **Completion quality**: Whether items bounced back after being delivered, or similar issues were opened indicating incomplete work.
- **PR activity**: Merged PRs, commit volume, cycle time, and code review engagement.
- **Rework detection**: Work items requiring multiple PRs to the same repository, signaling potential rework.

## Features

- **Single Report Mode**: Generate a detailed report for one developer over a date range.
- **Compare Mode**: Compare metrics side-by-side for 2-3 developers with bar charts.
- **Activity Timeline**: Line chart showing daily PR merge activity over the report period.
- **Repos Contributed To**: Collapsible panel showing PR aggregation per repository.
- **Work Item Type Breakdown**: Table view grouped by type and priority.
- **Quality Indicators**: Reopen rate with threshold-based ratings (Excellent/Good/Needs Improvement/Concerning).
- **Rework PRs**: Detection of work items with multiple PRs to the same repository.
- **Settings Page**: Configure ADO teams and work item types via the UI (no restart required).
- **Date Presets**: Quick-select buttons for Q1-Q4, half-years, YTD, and last 30 days.

## Architecture

- **Frontend**: Svelte + Vite + Chart.js (served via nginx in production)
- **Backend**: Go REST API that connects to Azure DevOps via PAT
- **Deployment**: Helm chart targeting AKS (Azure Kubernetes Service)

## Local Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- An Azure DevOps PAT with Work Items and Code (Git) read access

### Setup

1. Copy the environment file:
   ```bash
   cp .env.example .env
   ```
   Fill in your ADO organization URL, PAT, and project name.

2. Start the backend:
   ```bash
   cd backend
   go run ./cmd/server
   ```
   The backend runs on `http://localhost:8090` (configurable via `SERVER_PORT` in `.env`).

3. Start the frontend:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   The frontend runs on `http://localhost:5180` and proxies API calls to the backend.

### Docker Compose

```bash
docker-compose up --build
```

Frontend: `http://localhost:3000` | Backend API: `http://localhost:8090`

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ADO_ORG_URL` | Azure DevOps organization URL (e.g., `https://dev.azure.com/your-org`) |
| `ADO_PAT` | Personal Access Token with Work Items and Code read access |
| `ADO_PROJECT` | Azure DevOps project name |
| `ADO_TEAMS` | Comma-separated team names (initial seed for settings, optional) |
| `SERVER_PORT` | Backend port (default: `8080`) |

### Settings Page

Teams and work item types can be configured at runtime via the Settings page (gear icon in the header). Changes take effect immediately without restarting the server.

- **ADO Teams**: Which teams to load developers from. Leave empty for all project teams.
- **Work Item Types**: Which work item types to include in reports (default: Bug, Task).

Settings are persisted to a `settings.json` file alongside the server binary.

## Deployment (AKS)

See [helm/ado-reporter/README.md](helm/ado-reporter/README.md) for Helm chart usage.

```bash
helm install delivery-pulse ./helm/ado-reporter \
  --namespace your-namespace \
  --set config.adoOrgUrl="https://dev.azure.com/your-org" \
  --set config.adoProject="your-project" \
  --set secrets.adoPat="your-pat"
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/developers` | List developers from configured teams (cached) |
| GET | `/api/report?developer={email}&from={date}&to={date}` | Developer performance report |
| GET | `/api/workitems?developer={email}&from={date}&to={date}` | Detailed work items list |
| GET | `/api/settings` | Get current application settings |
| PUT | `/api/settings` | Update application settings (teams, work item types) |

## Metrics Reference

### Work Item Metrics
- **Total Work Items** — Items ever assigned to the developer with activity in the date range
- **Completed** — Items the developer handed off (first reassignment)
- **Avg Days to Complete** — Average time from first assignment to first reassignment
- **Bounced Back** — Items returned to the developer after handoff
- **Reopen Rate** — Percentage of items that bounced back (thresholds: 0% Excellent, 1-10% Good, 11-25% Needs Improvement, >25% Concerning)

### Pull Request Metrics
- **PRs Merged** — Completed (merged) PRs in the date range
- **Total Commits** — Sum of commits across all PRs
- **Avg PR Cycle (days)** — Average time from PR creation to merge
- **Files Changed** — Total files modified across merged PRs
- **Actionable Comments** — Reviewer comments that resulted in code changes
- **Rework PRs** — Work items with multiple PRs to the same repository
