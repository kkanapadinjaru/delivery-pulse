package ado

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client interacts with Azure DevOps REST API using a PAT.
type Client struct {
	orgURL     string
	pat        string
	project    string
	teams      []string // configured team names to load developers from
	workItemTypes []string // configured work item types to query
	// developerIDs maps uniqueName (email) to their ADO identity GUID
	developerIDs map[string]string
	httpClient   *http.Client

	// Developer cache
	cachedDevelopers []string
	cacheValid       bool
	cacheMu          sync.RWMutex
}

// NewClient creates a new Azure DevOps API client.
func NewClient(orgURL, pat, project string) *Client {
	return &Client{
		orgURL:        strings.TrimRight(orgURL, "/"),
		pat:           pat,
		project:       project,
		workItemTypes: []string{"Bug"},
		developerIDs:  make(map[string]string),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetTeams configures which team names to use for loading developers.
// If empty, all teams in the project are used.
// Invalidates the developer cache.
func (c *Client) SetTeams(teams []string) {
	c.teams = teams
	c.InvalidateDeveloperCache()
}

// InvalidateDeveloperCache forces the next GetDevelopers call to fetch fresh data.
func (c *Client) InvalidateDeveloperCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cacheValid = false
	c.cachedDevelopers = nil
}

// SetWorkItemTypes configures which work item types to query.
// If empty, defaults to ["Bug"].
func (c *Client) SetWorkItemTypes(types []string) {
	if len(types) == 0 {
		c.workItemTypes = []string{"Bug"}
	} else {
		c.workItemTypes = types
	}
}

// WorkItem represents a simplified Azure DevOps work item.
type WorkItem struct {
	ID             int       `json:"id"`
	Title          string    `json:"title"`
	State          string    `json:"state"`
	AssignedTo     string    `json:"assignedTo"`
	Priority       int       `json:"priority"`
	CreatedDate    time.Time `json:"createdDate"`
	ChangedDate    time.Time `json:"changedDate"`
	ResolvedDate   *time.Time `json:"resolvedDate,omitempty"`
	ClosedDate     *time.Time `json:"closedDate,omitempty"`
	AssignedDate   *time.Time `json:"assignedDate,omitempty"`
	ReassignedDate *time.Time `json:"reassignedDate,omitempty"`
	WorkItemType   string    `json:"workItemType"`
	Tags           string    `json:"tags"`
	ReactivatedCount int     `json:"reactivatedCount"`
	History        []StateChange `json:"history,omitempty"`
}

// StateChange tracks state transitions for a work item.
type StateChange struct {
	Date      time.Time `json:"date"`
	FromState string    `json:"fromState"`
	ToState   string    `json:"toState"`
	ChangedBy string    `json:"changedBy"`
}

// DeveloperReport contains the performance summary for a developer.
type DeveloperReport struct {
	Developer        string         `json:"developer"`
	From             string         `json:"from"`
	To               string         `json:"to"`
	AdoBaseUrl       string         `json:"adoBaseUrl"`
	TotalWorkedOn    int            `json:"totalWorkedOn"`
	Resolved         int            `json:"resolved"`
	Reopened         int            `json:"reopened"`
	ReopenRate       float64        `json:"reopenRate"`
	Priority1Count   int            `json:"priority1Count"`
	Priority2Count   int            `json:"priority2Count"`
	AvgResolutionDays float64       `json:"avgResolutionDays"`
	SimilarBugs      []SimilarBug   `json:"similarBugs"`
	WorkItems        []WorkItem     `json:"workItems"`
	PRMetrics        *PRMetrics     `json:"prMetrics,omitempty"`
	PRDetails        []PRDetail     `json:"prDetails,omitempty"`
}

// PRDetail represents a single pull request with its linked work item IDs.
type PRDetail struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Status        string    `json:"status"`
	Repository    string    `json:"repository"`
	CreatedDate   string    `json:"createdDate"`
	ClosedDate    string    `json:"closedDate,omitempty"`
	LinkedWorkItemIDs []int `json:"linkedWorkItemIds,omitempty"`
}

// PRMetrics contains pull request statistics for a developer.
type PRMetrics struct {
	PRsRaised          int     `json:"prsRaised"`
	PRsMerged          int     `json:"prsMerged"`
	TotalCommits       int     `json:"totalCommits"`
	AvgPRCycleDays     float64 `json:"avgPRCycleDays"`
	FilesChanged       int     `json:"filesChanged"`
	ActionableComments int     `json:"actionableComments"`
}

// SimilarBug represents a bug that was opened similar to one previously fixed.
type SimilarBug struct {
	OriginalID    int    `json:"originalId"`
	OriginalTitle string `json:"originalTitle"`
	NewID         int    `json:"newId"`
	NewTitle      string `json:"newTitle"`
	Similarity    string `json:"similarity"`
}

// wiqlResponse represents the WIQL query response.
type wiqlResponse struct {
	WorkItems []struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	} `json:"workItems"`
}

// workItemResponse represents a single work item from the API.
type workItemResponse struct {
	ID     int                    `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

// workItemBatchResponse represents a batch work item response.
type workItemBatchResponse struct {
	Count int                `json:"count"`
	Value []workItemResponse `json:"value"`
}

// updatesResponse represents work item updates (history).
type updatesResponse struct {
	Count int `json:"count"`
	Value []struct {
		RevisedDate string `json:"revisedDate"`
		Fields      map[string]struct {
			OldValue interface{} `json:"oldValue"`
			NewValue interface{} `json:"newValue"`
		} `json:"fields"`
		RevisedBy struct {
			DisplayName string `json:"displayName"`
			UniqueName  string `json:"uniqueName"`
		} `json:"revisedBy"`
	} `json:"value"`
}

// doRequest performs an authenticated request to Azure DevOps.
func (c *Client) doRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth("", c.pat)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// teamsResponse represents the list of teams in a project.
type teamsResponse struct {
	Count int `json:"count"`
	Value []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"value"`
}

// teamMembersResponse represents the list of members in a team.
type teamMembersResponse struct {
	Count int `json:"count"`
	Value []struct {
		IsTeamAdmin bool `json:"isTeamAdmin"`
		Identity    struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
			UniqueName  string `json:"uniqueName"`
			IsContainer bool   `json:"isContainer"`
		} `json:"identity"`
	} `json:"value"`
}

// GetDevelopers returns a list of team members from the configured teams.
// Uses uniqueName (email) for consistent formatting and filters out groups.
// Results are cached until InvalidateDeveloperCache is called.
func (c *Client) GetDevelopers() ([]string, error) {
	c.cacheMu.RLock()
	if c.cacheValid && c.cachedDevelopers != nil {
		result := make([]string, len(c.cachedDevelopers))
		copy(result, c.cachedDevelopers)
		c.cacheMu.RUnlock()
		return result, nil
	}
	c.cacheMu.RUnlock()

	developers, err := c.fetchDevelopersFromAPI()
	if err != nil {
		return nil, err
	}

	c.cacheMu.Lock()
	c.cachedDevelopers = developers
	c.cacheValid = true
	c.cacheMu.Unlock()

	return developers, nil
}

// fetchDevelopersFromAPI queries the ADO API for team members.
func (c *Client) fetchDevelopersFromAPI() ([]string, error) {
	// Get all teams in the project
	teamsURL := fmt.Sprintf("%s/_apis/projects/%s/teams?api-version=7.0", c.orgURL, c.project)
	data, err := c.doRequest("GET", teamsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching teams: %w", err)
	}

	var teams teamsResponse
	if err := json.Unmarshal(data, &teams); err != nil {
		return nil, fmt.Errorf("parsing teams response: %w", err)
	}

	// Filter to configured teams if specified
	var targetTeams []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if len(c.teams) > 0 {
		for _, team := range teams.Value {
			for _, configured := range c.teams {
				if strings.EqualFold(team.Name, configured) {
					targetTeams = append(targetTeams, team)
					break
				}
			}
		}
	} else {
		targetTeams = teams.Value
	}

	// Get members from each target team
	developers := make(map[string]bool)
	for _, team := range targetTeams {
		membersURL := fmt.Sprintf("%s/_apis/projects/%s/teams/%s/members?api-version=7.0", c.orgURL, c.project, team.ID)
		memberData, err := c.doRequest("GET", membersURL, nil)
		if err != nil {
			continue // skip teams we can't access
		}

		var members teamMembersResponse
		if err := json.Unmarshal(memberData, &members); err != nil {
			continue
		}

		for _, m := range members.Value {
			// Skip groups (containers) — only include individual users
			if m.Identity.IsContainer {
				continue
			}
			if m.Identity.UniqueName != "" {
				developers[m.Identity.UniqueName] = true
				// Cache the identity GUID for PR queries
				if m.Identity.ID != "" {
					c.developerIDs[strings.ToLower(m.Identity.UniqueName)] = m.Identity.ID
				}
			}
		}
	}

	result := make([]string, 0, len(developers))
	for dev := range developers {
		result = append(result, dev)
	}
	return result, nil
}

// GetDeveloperReport generates a performance report for a developer in a date range.
func (c *Client) GetDeveloperReport(developer, from, to string) (*DeveloperReport, error) {
	workItems, err := c.getWorkItemsForDeveloper(developer, from, to)
	if err != nil {
		return nil, err
	}

	report := &DeveloperReport{
		Developer:     developer,
		From:          from,
		To:            to,
		AdoBaseUrl:    fmt.Sprintf("%s/%s", c.orgURL, c.project),
		TotalWorkedOn: len(workItems),
		WorkItems:     workItems,
	}

	var totalResolutionDays float64
	resolvedCount := 0

	for i := range workItems {
		wi := &workItems[i]

		switch wi.Priority {
		case 1:
			report.Priority1Count++
		case 2:
			report.Priority2Count++
		}

		// A bug is considered "fixed" by the developer when they reassigned it away
		if wi.ReassignedDate != nil {
			report.Resolved++
		}

		if wi.ReactivatedCount > 0 {
			report.Reopened++
		}

		// Avg resolution time = time from when dev was assigned to when they reassigned it
		if wi.ReassignedDate != nil && wi.AssignedDate != nil {
			days := wi.ReassignedDate.Sub(*wi.AssignedDate).Hours() / 24
			totalResolutionDays += days
			resolvedCount++
		}
	}

	if report.TotalWorkedOn > 0 {
		report.ReopenRate = float64(report.Reopened) / float64(report.TotalWorkedOn) * 100
	}
	if resolvedCount > 0 {
		report.AvgResolutionDays = totalResolutionDays / float64(resolvedCount)
	}

	// Find similar bugs (basic title-matching heuristic)
	report.SimilarBugs = c.findSimilarBugs(workItems)

	// Fetch PR metrics and details
	prMetrics, prDetails, err := c.GetPRMetrics(developer, from, to)
	if err != nil {
		log.Printf("Warning: could not fetch PR metrics for %s: %v", developer, err)
	} else {
		report.PRMetrics = prMetrics
		report.PRDetails = prDetails
	}

	return report, nil
}

// GetWorkItems returns detailed work items for a developer in a date range.
func (c *Client) GetWorkItems(developer, from, to string) ([]WorkItem, error) {
	return c.getWorkItemsForDeveloper(developer, from, to)
}

// getWorkItemsForDeveloper fetches work items that a developer worked on in the given period.
// Uses WIQL with "EVER" and "ASOF" to find items even if the developer is no longer the assignee.
func (c *Client) getWorkItemsForDeveloper(developer, from, to string) ([]WorkItem, error) {
	// Build work item type filter
	typeFilter := "'Bug'" // default fallback
	if len(c.workItemTypes) > 0 {
		types := make([]string, len(c.workItemTypes))
		for i, t := range c.workItemTypes {
			types[i] = fmt.Sprintf("'%s'", t)
		}
		typeFilter = strings.Join(types, ", ")
	}

	// Query for work items that were ever assigned to this developer and changed in the date range
	wiql := fmt.Sprintf(`SELECT [System.Id] FROM WorkItems 
		WHERE [System.WorkItemType] IN (%s) 
		AND [System.ChangedDate] >= '%s' 
		AND [System.ChangedDate] <= '%s'
		AND [System.AssignedTo] EVER '%s'
		ORDER BY [System.ChangedDate] DESC`, typeFilter, from, to, developer)

	url := fmt.Sprintf("%s/%s/_apis/wit/wiql?api-version=7.0", c.orgURL, c.project)
	body := strings.NewReader(fmt.Sprintf(`{"query": %q}`, wiql))

	data, err := c.doRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	var result wiqlResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing WIQL response: %w", err)
	}

	ids := make([]int, 0, len(result.WorkItems))
	for _, wi := range result.WorkItems {
		ids = append(ids, wi.ID)
	}

	if len(ids) == 0 {
		return []WorkItem{}, nil
	}

	workItems, err := c.getWorkItemsBatch(ids)
	if err != nil {
		return nil, err
	}

	// Enrich with history (reactivation count, assignment dates) — parallel with concurrency limit
	sem := make(chan struct{}, 10) // max 10 concurrent ADO API calls
	var wg sync.WaitGroup
	for i := range workItems {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			updates, err := c.getWorkItemUpdates(workItems[idx].ID)
			if err != nil {
				return
			}
			workItems[idx].History = extractStateChanges(updates)
			info := extractAssignmentInfo(updates, developer)
			workItems[idx].AssignedDate = info.FirstAssignedDate
			workItems[idx].ReassignedDate = info.FirstReassignedDate
			workItems[idx].ReactivatedCount = info.BounceBackCount
		}(i)
	}
	wg.Wait()

	return workItems, nil
}

// getWorkItemsBatch fetches work items by IDs in batches.
func (c *Client) getWorkItemsBatch(ids []int) ([]WorkItem, error) {
	var allItems []WorkItem
	batchSize := 200

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]

		idStrs := make([]string, len(batch))
		for j, id := range batch {
			idStrs[j] = fmt.Sprintf("%d", id)
		}

		fields := "System.Id,System.Title,System.State,System.AssignedTo,Microsoft.VSTS.Common.Priority,System.CreatedDate,System.ChangedDate,Microsoft.VSTS.Common.ResolvedDate,Microsoft.VSTS.Common.ClosedDate,System.WorkItemType,System.Tags"
		url := fmt.Sprintf("%s/%s/_apis/wit/workitems?ids=%s&fields=%s&api-version=7.0",
			c.orgURL, c.project, strings.Join(idStrs, ","), fields)

		data, err := c.doRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		var resp workItemBatchResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("parsing batch response: %w", err)
		}

		for _, item := range resp.Value {
			wi := parseWorkItem(item)
			allItems = append(allItems, wi)
		}
	}

	return allItems, nil
}

// getWorkItemUpdates fetches the revision history for a work item.
func (c *Client) getWorkItemUpdates(id int) ([]struct {
	RevisedDate string `json:"revisedDate"`
	Fields      map[string]struct {
		OldValue interface{} `json:"oldValue"`
		NewValue interface{} `json:"newValue"`
	} `json:"fields"`
	RevisedBy struct {
		DisplayName string `json:"displayName"`
		UniqueName  string `json:"uniqueName"`
	} `json:"revisedBy"`
}, error) {
	url := fmt.Sprintf("%s/%s/_apis/wit/workitems/%d/updates?api-version=7.0", c.orgURL, c.project, id)

	data, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var resp updatesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing updates response: %w", err)
	}

	return resp.Value, nil
}

// assignmentInfo holds the assignment analysis results for a work item.
type assignmentInfo struct {
	FirstAssignedDate  *time.Time // When the developer was first assigned this bug
	FirstReassignedDate *time.Time // When the developer first handed it off
	BounceBackCount    int        // Times the bug came back to the developer after they handed it off
}

// extractAssignmentInfo analyzes the work item history to determine:
// - When the developer was first assigned the bug
// - When they first reassigned it away (indicating they finished the fix)
// - How many times the bug bounced back to them (indicating tester rejected the fix)
func extractAssignmentInfo(updates []struct {
	RevisedDate string `json:"revisedDate"`
	Fields      map[string]struct {
		OldValue interface{} `json:"oldValue"`
		NewValue interface{} `json:"newValue"`
	} `json:"fields"`
	RevisedBy struct {
		DisplayName string `json:"displayName"`
		UniqueName  string `json:"uniqueName"`
	} `json:"revisedBy"`
}, developer string) assignmentInfo {
	info := assignmentInfo{}
	hasBeenAssigned := false
	hasBeenReassigned := false

	for _, update := range updates {
		assignedToField, ok := update.Fields["System.AssignedTo"]
		if !ok {
			continue
		}

		newName := extractDisplayName(assignedToField.NewValue)
		oldName := extractDisplayName(assignedToField.OldValue)
		newUnique := extractUniqueName(assignedToField.NewValue)
		oldUnique := extractUniqueName(assignedToField.OldValue)

		isDeveloperNew := strings.EqualFold(newName, developer) || strings.EqualFold(newUnique, developer)
		isDeveloperOld := strings.EqualFold(oldName, developer) || strings.EqualFold(oldUnique, developer)

		if isDeveloperNew {
			// Developer was assigned this bug
			t, err := time.Parse(time.RFC3339, update.RevisedDate)
			if err != nil {
				continue
			}

			if !hasBeenAssigned {
				// First time assigned
				info.FirstAssignedDate = &t
				hasBeenAssigned = true
			} else if hasBeenReassigned {
				// Bug came back to the developer after they already handed it off
				info.BounceBackCount++
			}
		} else if isDeveloperOld && (newName != "" || newUnique != "") {
			// Developer reassigned the bug to someone else
			t, err := time.Parse(time.RFC3339, update.RevisedDate)
			if err != nil {
				continue
			}

			if !hasBeenReassigned {
				// First time they handed it off
				info.FirstReassignedDate = &t
				hasBeenReassigned = true
			}
		}
	}

	return info
}

// extractDisplayName gets the display name from an AssignedTo field value,
// which can be a string or a map with a "displayName" key.
func extractDisplayName(val interface{}) string {
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	if m, ok := val.(map[string]interface{}); ok {
		if name, ok := m["displayName"].(string); ok {
			return name
		}
	}
	return ""
}

// extractUniqueName gets the uniqueName (email) from an AssignedTo field value.
func extractUniqueName(val interface{}) string {
	if val == nil {
		return ""
	}
	if m, ok := val.(map[string]interface{}); ok {
		if name, ok := m["uniqueName"].(string); ok {
			return name
		}
	}
	return ""
}

// extractStateChanges builds a timeline of state changes from updates.
func extractStateChanges(updates []struct {
	RevisedDate string `json:"revisedDate"`
	Fields      map[string]struct {
		OldValue interface{} `json:"oldValue"`
		NewValue interface{} `json:"newValue"`
	} `json:"fields"`
	RevisedBy struct {
		DisplayName string `json:"displayName"`
		UniqueName  string `json:"uniqueName"`
	} `json:"revisedBy"`
}) []StateChange {
	var changes []StateChange
	for _, update := range updates {
		if stateField, ok := update.Fields["System.State"]; ok {
			oldState, _ := stateField.OldValue.(string)
			newState, _ := stateField.NewValue.(string)
			if oldState != "" || newState != "" {
				t, _ := time.Parse(time.RFC3339, update.RevisedDate)
				changes = append(changes, StateChange{
					Date:      t,
					FromState: oldState,
					ToState:   newState,
					ChangedBy: update.RevisedBy.DisplayName,
				})
			}
		}
	}
	return changes
}

// parseWorkItem converts an API response to our WorkItem struct.
func parseWorkItem(item workItemResponse) WorkItem {
	wi := WorkItem{
		ID: item.ID,
	}

	if v, ok := item.Fields["System.Title"].(string); ok {
		wi.Title = v
	}
	if v, ok := item.Fields["System.State"].(string); ok {
		wi.State = v
	}
	if v, ok := item.Fields["System.WorkItemType"].(string); ok {
		wi.WorkItemType = v
	}
	if v, ok := item.Fields["System.Tags"].(string); ok {
		wi.Tags = v
	}
	if v, ok := item.Fields["Microsoft.VSTS.Common.Priority"].(float64); ok {
		wi.Priority = int(v)
	}
	if v, ok := item.Fields["System.AssignedTo"].(map[string]interface{}); ok {
		if name, ok := v["displayName"].(string); ok {
			wi.AssignedTo = name
		}
	}
	if v, ok := item.Fields["System.CreatedDate"].(string); ok {
		t, _ := time.Parse(time.RFC3339, v)
		wi.CreatedDate = t
	}
	if v, ok := item.Fields["System.ChangedDate"].(string); ok {
		t, _ := time.Parse(time.RFC3339, v)
		wi.ChangedDate = t
	}
	if v, ok := item.Fields["Microsoft.VSTS.Common.ResolvedDate"].(string); ok {
		t, _ := time.Parse(time.RFC3339, v)
		wi.ResolvedDate = &t
	}
	if v, ok := item.Fields["Microsoft.VSTS.Common.ClosedDate"].(string); ok {
		t, _ := time.Parse(time.RFC3339, v)
		wi.ClosedDate = &t
	}

	return wi
}

// findSimilarBugs uses a basic title similarity check to find bugs that may be related.
func (c *Client) findSimilarBugs(workItems []WorkItem) []SimilarBug {
	var similar []SimilarBug

	for i := 0; i < len(workItems); i++ {
		for j := i + 1; j < len(workItems); j++ {
			if titleSimilarity(workItems[i].Title, workItems[j].Title) > 0.6 {
				// Determine which is original vs new based on creation date
				orig := workItems[i]
				new := workItems[j]
				if workItems[j].CreatedDate.Before(workItems[i].CreatedDate) {
					orig = workItems[j]
					new = workItems[i]
				}
				similar = append(similar, SimilarBug{
					OriginalID:    orig.ID,
					OriginalTitle: orig.Title,
					NewID:         new.ID,
					NewTitle:      new.Title,
					Similarity:    "title-match",
				})
			}
		}
	}

	return similar
}

// titleSimilarity calculates a simple Jaccard similarity between two titles.
func titleSimilarity(a, b string) float64 {
	wordsA := strings.Fields(strings.ToLower(a))
	wordsB := strings.Fields(strings.ToLower(b))

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0
	}

	setA := make(map[string]bool)
	for _, w := range wordsA {
		setA[w] = true
	}

	setB := make(map[string]bool)
	for _, w := range wordsB {
		setB[w] = true
	}

	intersection := 0
	for w := range setA {
		if setB[w] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// pullRequestsResponse represents the list of pull requests from the API.
type pullRequestsResponse struct {
	Count int          `json:"count"`
	Value []pullRequest `json:"value"`
}

// pullRequest represents a single pull request from the API.
type pullRequest struct {
	PullRequestID int    `json:"pullRequestId"`
	Title         string `json:"title"`
	Status        string `json:"status"`
	CreatedBy     struct {
		DisplayName string `json:"displayName"`
		UniqueName  string `json:"uniqueName"`
	} `json:"createdBy"`
	CreationDate  string `json:"creationDate"`
	ClosedDate    string `json:"closedDate"`
	Repository    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"repository"`
	LastMergeCommit *struct {
		CommitID string `json:"commitId"`
	} `json:"lastMergeCommit"`
}

// prCommitsResponse represents the commits on a pull request.
type prCommitsResponse struct {
	Count int `json:"count"`
	Value []struct {
		CommitID string `json:"commitId"`
	} `json:"value"`
}

// prIterationsResponse represents the iterations (pushes) on a PR.
type prIterationsResponse struct {
	Count int `json:"count"`
	Value []struct {
		ID int `json:"id"`
	} `json:"value"`
}

// prThreadsResponse represents all threads (comments) on a PR.
type prThreadsResponse struct {
	Count int        `json:"count"`
	Value []prThread `json:"value"`
}

// prThread represents a single comment thread on a PR.
type prThread struct {
	ID       int `json:"id"`
	Status   string `json:"status"`
	Comments []struct {
		Author struct {
			DisplayName string `json:"displayName"`
			UniqueName  string `json:"uniqueName"`
		} `json:"author"`
		CommentType   string `json:"commentType"`
		PublishedDate string `json:"publishedDate"`
	} `json:"comments"`
	ThreadContext *struct {
		FilePath string `json:"filePath"`
	} `json:"threadContext"`
	PullRequestThreadContext *struct {
		IterationContext struct {
			FirstComparingIteration  int `json:"firstComparingIteration"`
			SecondComparingIteration int `json:"secondComparingIteration"`
		} `json:"iterationContext"`
	} `json:"pullRequestThreadContext"`
}

// GetPRMetrics fetches pull request metrics for a developer within a date range.
// Also returns detailed PR info with linked work items.
func (c *Client) GetPRMetrics(developer, from, to string) (*PRMetrics, []PRDetail, error) {
	// Parse dates for comparison
	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing from date: %w", err)
	}
	toTime, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing to date: %w", err)
	}
	// Include the full end day
	toTime = toTime.Add(24*time.Hour - time.Second)

	// Get PRs for this developer in the date range (server-side filtered)
	prs, err := c.getProjectPullRequests(developer, fromTime, toTime)
	if err != nil {
		return nil, nil, err
	}

	metrics := &PRMetrics{}
	var totalCycleDays float64
	cycleCount := 0
	details := make([]PRDetail, len(prs))

	// First pass: compute metrics from PR metadata (no extra API calls)
	for i, pr := range prs {
		metrics.PRsRaised++

		details[i] = PRDetail{
			ID:          pr.PullRequestID,
			Title:       pr.Title,
			Status:      pr.Status,
			Repository:  pr.Repository.Name,
			CreatedDate: pr.CreationDate,
			ClosedDate:  pr.ClosedDate,
		}

		if pr.Status == "completed" {
			metrics.PRsMerged++

			if pr.ClosedDate != "" {
				createdTime, err := time.Parse(time.RFC3339, pr.CreationDate)
				closedTime, err2 := time.Parse(time.RFC3339, pr.ClosedDate)
				if err == nil && err2 == nil {
					cycleDays := closedTime.Sub(createdTime).Hours() / 24
					totalCycleDays += cycleDays
					cycleCount++
				}
			}
		}
	}

	if cycleCount > 0 {
		metrics.AvgPRCycleDays = totalCycleDays / float64(cycleCount)
	}

	// Second pass: fetch per-PR details in parallel (commits, files, comments, linked items)
	type prResult struct {
		idx              int
		commits          int
		filesChanged     int
		actionable       int
		linkedWorkItems  []int
	}

	results := make([]prResult, len(prs))
	sem := make(chan struct{}, 10) // max 10 concurrent calls
	var wg sync.WaitGroup

	for i, pr := range prs {
		wg.Add(1)
		go func(idx int, pr pullRequest) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			r := prResult{idx: idx}

			// Commits
			if commits, err := c.getPRCommitCount(pr.Repository.ID, pr.PullRequestID); err == nil {
				r.commits = commits
			}

			// Files changed (only for merged PRs)
			if pr.Status == "completed" {
				r.filesChanged = c.getPRFilesChanged(pr.Repository.ID, pr.PullRequestID)
			}

			// Actionable comments
			r.actionable = c.getActionableComments(pr.Repository.ID, pr.PullRequestID, developer)

			// Linked work items
			r.linkedWorkItems = c.getPRLinkedWorkItems(pr.Repository.ID, pr.PullRequestID)

			results[idx] = r
		}(i, pr)
	}
	wg.Wait()

	// Aggregate parallel results
	for _, r := range results {
		metrics.TotalCommits += r.commits
		metrics.FilesChanged += r.filesChanged
		metrics.ActionableComments += r.actionable
		if len(r.linkedWorkItems) > 0 {
			details[r.idx].LinkedWorkItemIDs = r.linkedWorkItems
		}
	}

	return metrics, details, nil
}

// getProjectPullRequests fetches pull requests created by a developer within a date range.
// Uses searchCriteria.creatorId and minTime/maxTime for server-side filtering.
// Makes separate calls for completed and abandoned PRs (skips active/draft which are in-progress).
func (c *Client) getProjectPullRequests(developer string, from, to time.Time) ([]pullRequest, error) {
	// Resolve the developer email to their ADO identity GUID
	creatorID := c.developerIDs[strings.ToLower(developer)]
	if creatorID == "" {
		return nil, fmt.Errorf("could not resolve identity GUID for %s", developer)
	}

	// Format dates for the API (ISO 8601)
	minTime := from.Format("2006-01-02T15:04:05Z")
	maxTime := to.Format("2006-01-02T15:04:05Z")

	var allPRs []pullRequest

	// Fetch completed and abandoned PRs (the two terminal states)
	for _, status := range []string{"completed", "abandoned"} {
		skip := 0
		top := 100
		for {
			url := fmt.Sprintf("%s/%s/_apis/git/pullrequests?searchCriteria.creatorId=%s&searchCriteria.status=%s&searchCriteria.minTime=%s&searchCriteria.maxTime=%s&$top=%d&$skip=%d&api-version=7.0",
				c.orgURL, c.project, creatorID, status, minTime, maxTime, top, skip)

			data, err := c.doRequest("GET", url, nil)
			if err != nil {
				return nil, fmt.Errorf("fetching pull requests (status=%s): %w", status, err)
			}

			var resp pullRequestsResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return nil, fmt.Errorf("parsing PR response: %w", err)
			}

			allPRs = append(allPRs, resp.Value...)

			if len(resp.Value) < top {
				break
			}
			skip += top
		}
	}

	return allPRs, nil
}

// getPRCommitCount returns the number of commits in a pull request.
func (c *Client) getPRCommitCount(repoID string, prID int) (int, error) {
	url := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullRequests/%d/commits?api-version=7.0",
		c.orgURL, c.project, repoID, prID)

	data, err := c.doRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	var resp prCommitsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, err
	}

	return resp.Count, nil
}

// getPRFilesChanged returns the number of files changed in a PR using the last iteration changes.
func (c *Client) getPRFilesChanged(repoID string, prID int) int {
	// Get PR iterations
	url := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullRequests/%d/iterations?api-version=7.0",
		c.orgURL, c.project, repoID, prID)

	data, err := c.doRequest("GET", url, nil)
	if err != nil {
		return 0
	}

	var iters prIterationsResponse
	if err := json.Unmarshal(data, &iters); err != nil || iters.Count == 0 {
		return 0
	}

	// Get changes for the last iteration (the final state of the PR diff)
	lastIterID := iters.Value[iters.Count-1].ID
	changesURL := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullRequests/%d/iterations/%d/changes?api-version=7.0",
		c.orgURL, c.project, repoID, prID, lastIterID)

	changesData, err := c.doRequest("GET", changesURL, nil)
	if err != nil {
		return 0
	}

	var changes struct {
		ChangeEntries []struct {
			ChangeTrackingID int `json:"changeTrackingId"`
		} `json:"changeEntries"`
	}
	if err := json.Unmarshal(changesData, &changes); err != nil {
		return 0
	}

	return len(changes.ChangeEntries)
}

// getActionableComments counts review comments from other people that were followed by subsequent
// iterations (pushes/commits), indicating the developer had to make code changes to address them.
func (c *Client) getActionableComments(repoID string, prID int, prAuthor string) int {
	// Get total iterations for this PR
	itersURL := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullRequests/%d/iterations?api-version=7.0",
		c.orgURL, c.project, repoID, prID)

	itersData, err := c.doRequest("GET", itersURL, nil)
	if err != nil {
		return 0
	}

	var iters prIterationsResponse
	if err := json.Unmarshal(itersData, &iters); err != nil || iters.Count <= 1 {
		// If there's only 1 iteration, no commits came after any comment
		return 0
	}
	totalIterations := iters.Count

	// Get all threads for this PR
	threadsURL := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullRequests/%d/threads?api-version=7.0",
		c.orgURL, c.project, repoID, prID)

	threadsData, err := c.doRequest("GET", threadsURL, nil)
	if err != nil {
		return 0
	}

	var threads prThreadsResponse
	if err := json.Unmarshal(threadsData, &threads); err != nil {
		return 0
	}

	actionable := 0
	for _, thread := range threads.Value {
		// Skip system-generated threads and threads without code context
		if len(thread.Comments) == 0 {
			continue
		}

		// The first comment in the thread is the initiating review comment
		firstComment := thread.Comments[0]

		// Skip system comments (status updates, vote changes, etc.)
		if firstComment.CommentType == "system" {
			continue
		}

		// Skip comments by the PR author themselves (self-comments aren't reviewer feedback)
		if strings.EqualFold(firstComment.Author.DisplayName, prAuthor) || strings.EqualFold(firstComment.Author.UniqueName, prAuthor) {
			continue
		}

		// Check if this thread is on a code file (has thread context)
		// and was made on an iteration before the final one
		if thread.PullRequestThreadContext != nil {
			commentIteration := thread.PullRequestThreadContext.IterationContext.SecondComparingIteration
			if commentIteration > 0 && commentIteration < totalIterations {
				// There were subsequent pushes after this comment — it was actionable
				actionable++
			}
		} else if thread.ThreadContext != nil {
			// Has file context but no iteration context — count if thread was resolved/fixed
			if thread.Status == "fixed" || thread.Status == "closed" {
				actionable++
			}
		}
	}

	return actionable
}

// getPRLinkedWorkItems fetches the work item IDs linked to a pull request.
func (c *Client) getPRLinkedWorkItems(repoID string, prID int) []int {
	url := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullRequests/%d/workitems?api-version=7.0",
		c.orgURL, c.project, repoID, prID)

	data, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil
	}

	var resp struct {
		Count int `json:"count"`
		Value []struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"value"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil
	}

	var ids []int
	for _, wi := range resp.Value {
		// The ID field in the PR work items response is a string
		var id int
		if _, err := fmt.Sscanf(wi.ID, "%d", &id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}
