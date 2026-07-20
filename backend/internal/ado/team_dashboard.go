package ado

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// TeamDashboard contains aggregate team-level metrics.
type TeamDashboard struct {
	From                  string              `json:"from"`
	To                    string              `json:"to"`
	TotalMembers          int                 `json:"totalMembers"`
	TotalCompleted        int                 `json:"totalCompleted"`
	TotalWIP              int                 `json:"totalWip"`
	AvgCycleTimeDays      float64             `json:"avgCycleTimeDays"`
	ThroughputTrend       []ThroughputWeek    `json:"throughputTrend"`
	CycleTimeDistribution []CycleTimeBucket   `json:"cycleTimeDistribution"`
	WIPByMember           []MemberWIP         `json:"wipByMember"`
}

// CycleTimeBucket represents a histogram bucket for cycle time distribution.
type CycleTimeBucket struct {
	Label string `json:"label"` // e.g. "0-2 days", "3-5 days"
	Count int    `json:"count"`
}

// MemberWIP shows how many items are currently in-progress for a team member.
type MemberWIP struct {
	Developer string `json:"developer"`
	Count     int    `json:"count"`
}

// GetTeamDashboard aggregates metrics across all developers in configured teams.
func (c *Client) GetTeamDashboard(from, to string) (*TeamDashboard, error) {
	developers, err := c.GetDevelopers()
	if err != nil {
		return nil, fmt.Errorf("fetching developers: %w", err)
	}

	dashboard := &TeamDashboard{
		From:         from,
		To:           to,
		TotalMembers: len(developers),
	}

	// Fetch all team work items via a single WIQL query (no developer filter)
	allItems, err := c.getTeamWorkItems(from, to)
	if err != nil {
		return nil, fmt.Errorf("fetching team work items: %w", err)
	}

	// Compute metrics
	var totalCycleDays float64
	var cycleCount int

	for _, wi := range allItems {
		// Completed = has a reassigned date (developer handed it off)
		if wi.ReassignedDate != nil {
			dashboard.TotalCompleted++

			// Cycle time = assigned to reassigned
			if wi.AssignedDate != nil {
				days := wi.ReassignedDate.Sub(*wi.AssignedDate).Hours() / 24
				if days >= 0 {
					totalCycleDays += days
					cycleCount++
				}
			}
		}

		// WIP = items not yet completed and not closed/resolved
		if wi.ReassignedDate == nil && wi.State != "Closed" && wi.State != "Resolved" && wi.State != "Done" {
			dashboard.TotalWIP++
		}
	}

	if cycleCount > 0 {
		dashboard.AvgCycleTimeDays = math.Round(totalCycleDays/float64(cycleCount)*100) / 100
	}

	// Throughput trend (weekly buckets of completed items across team)
	dashboard.ThroughputTrend = computeThroughputTrend(allItems, from, to)

	// Cycle time distribution histogram
	dashboard.CycleTimeDistribution = computeCycleTimeDistribution(allItems)

	// WIP by member (based on current assignee, filtered to configured developers)
	dashboard.WIPByMember = computeWIPByMember(allItems, developers)

	return dashboard, nil
}

// getTeamWorkItems fetches all work items in the configured area paths/types
// that had activity in the date range — not scoped to a single developer.
func (c *Client) getTeamWorkItems(from, to string) ([]WorkItem, error) {
	// Build work item type filter
	typeFilter := "'Bug'"
	if len(c.workItemTypes) > 0 {
		types := make([]string, len(c.workItemTypes))
		for i, t := range c.workItemTypes {
			types[i] = fmt.Sprintf("'%s'", t)
		}
		typeFilter = strings.Join(types, ", ")
	}

	// Build area path filter
	var areaPathClause string
	if len(c.areaPaths) > 0 {
		parts := make([]string, len(c.areaPaths))
		for i, ap := range c.areaPaths {
			parts[i] = fmt.Sprintf("[System.AreaPath] = '%s'", ap)
		}
		areaPathClause = fmt.Sprintf("AND (%s)", strings.Join(parts, " OR "))
	}

	// Build activity filter
	var activityClause string
	if len(c.activities) > 0 {
		acts := make([]string, len(c.activities))
		for i, a := range c.activities {
			acts[i] = fmt.Sprintf("'%s'", a)
		}
		activityClause = fmt.Sprintf("AND [Microsoft.VSTS.Common.Activity] IN (%s)", strings.Join(acts, ", "))
	}

	wiql := fmt.Sprintf(`SELECT [System.Id] FROM WorkItems 
		WHERE [System.WorkItemType] IN (%s) 
		AND [System.ChangedDate] >= '%s' 
		AND [System.ChangedDate] <= '%s'
		%s
		%s
		ORDER BY [System.ChangedDate] DESC`, typeFilter, from, to, areaPathClause, activityClause)

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

	// Enrich with history for cycle time calculation
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	var enrichErrors []int
	var errMu sync.Mutex

	for i := range workItems {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			updates, err := c.getWorkItemUpdates(workItems[idx].ID)
			if err != nil {
				errMu.Lock()
				enrichErrors = append(enrichErrors, workItems[idx].ID)
				errMu.Unlock()
				return
			}
			workItems[idx].History = extractStateChanges(updates)
			// For team dashboard we need assignment info for any developer
			// Use AssignedTo (current) to find assignment dates
			info := extractFirstAssignment(updates)
			workItems[idx].AssignedDate = info.assignedDate
			workItems[idx].ReassignedDate = info.reassignedDate
		}(i)
	}
	wg.Wait()

	if len(enrichErrors) > 0 {
		c.logger.Warn("partial team work item enrichment",
			"totalItems", len(workItems),
			"failedItems", len(enrichErrors),
		)
	}

	return workItems, nil
}

// firstAssignmentInfo holds the first assignment and reassignment for any developer.
type firstAssignmentInfo struct {
	assignedDate   *time.Time
	reassignedDate *time.Time
}

// extractFirstAssignment finds when the item was first assigned and first reassigned,
// regardless of which developer. Used for team-level cycle time.
func extractFirstAssignment(updates []struct {
	RevisedDate string `json:"revisedDate"`
	Fields      map[string]struct {
		OldValue interface{} `json:"oldValue"`
		NewValue interface{} `json:"newValue"`
	} `json:"fields"`
	RevisedBy struct {
		DisplayName string `json:"displayName"`
		UniqueName  string `json:"uniqueName"`
	} `json:"revisedBy"`
}) firstAssignmentInfo {
	info := firstAssignmentInfo{}

	for _, update := range updates {
		assignedToField, ok := update.Fields["System.AssignedTo"]
		if !ok {
			continue
		}

		newName := extractDisplayName(assignedToField.NewValue)
		oldName := extractDisplayName(assignedToField.OldValue)

		t, err := time.Parse(time.RFC3339, update.RevisedDate)
		if err != nil {
			continue
		}

		// First assignment: oldValue is empty/nil, newValue is someone
		if (oldName == "" && newName != "") && info.assignedDate == nil {
			info.assignedDate = &t
		}

		// First reassignment: both old and new have values (handed off)
		if oldName != "" && newName != "" && oldName != newName && info.reassignedDate == nil && info.assignedDate != nil {
			info.reassignedDate = &t
		}
	}

	return info
}

// computeCycleTimeDistribution creates histogram buckets of cycle times.
func computeCycleTimeDistribution(items []WorkItem) []CycleTimeBucket {
	buckets := []CycleTimeBucket{
		{Label: "0-2 days", Count: 0},
		{Label: "3-5 days", Count: 0},
		{Label: "6-10 days", Count: 0},
		{Label: "11-20 days", Count: 0},
		{Label: "21-30 days", Count: 0},
		{Label: "30+ days", Count: 0},
	}

	for _, wi := range items {
		if wi.ReassignedDate == nil || wi.AssignedDate == nil {
			continue
		}
		days := wi.ReassignedDate.Sub(*wi.AssignedDate).Hours() / 24
		if days < 0 {
			continue
		}

		switch {
		case days <= 2:
			buckets[0].Count++
		case days <= 5:
			buckets[1].Count++
		case days <= 10:
			buckets[2].Count++
		case days <= 20:
			buckets[3].Count++
		case days <= 30:
			buckets[4].Count++
		default:
			buckets[5].Count++
		}
	}

	return buckets
}

// computeWIPByMember counts in-progress items per current assignee.
// Only includes members in the developers list (the filtered allowlist).
// Matches by email (case-insensitive) and by derived display name (First Last from email).
func computeWIPByMember(items []WorkItem, developers []string) []MemberWIP {
	wipMap := make(map[string]int)

	for _, wi := range items {
		if wi.ReassignedDate == nil && wi.State != "Closed" && wi.State != "Resolved" && wi.State != "Done" {
			if wi.AssignedTo != "" {
				wipMap[wi.AssignedTo]++
			}
		}
	}

	if len(developers) == 0 {
		// No filter — return all
		result := make([]MemberWIP, 0, len(wipMap))
		for assignee, count := range wipMap {
			result = append(result, MemberWIP{Developer: assignee, Count: count})
		}
		return result
	}

	// Build allowed set: emails + derived display names from emails
	allowedEmails := make(map[string]bool)
	allowedNames := make(map[string]bool)
	for _, email := range developers {
		allowedEmails[strings.ToLower(email)] = true
		// Derive display name: "first.last@domain" -> "first last"
		local := strings.Split(strings.ToLower(email), "@")[0]
		name := strings.ReplaceAll(local, ".", " ")
		allowedNames[name] = true
	}

	result := make([]MemberWIP, 0)
	for assignee, count := range wipMap {
		lower := strings.ToLower(assignee)
		if allowedEmails[lower] || allowedNames[lower] {
			result = append(result, MemberWIP{Developer: assignee, Count: count})
		}
	}

	return result
}
