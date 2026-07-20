package ado

import (
	"math"
	"time"
)

// TimeInState represents the average days spent in each state across work items.
type TimeInState struct {
	State   string  `json:"state"`
	AvgDays float64 `json:"avgDays"`
}

// ThroughputWeek represents the number of items completed in a given week.
type ThroughputWeek struct {
	WeekStart string `json:"weekStart"` // ISO date (Monday of that week)
	Count     int    `json:"count"`
}

// PRSizeBuckets categorizes PRs by number of files changed.
type PRSizeBuckets struct {
	Small     int `json:"small"`     // <= SmallMax files
	Medium    int `json:"medium"`    // SmallMax < files <= MediumMax
	Large     int `json:"large"`     // > MediumMax files
	SmallMax  int `json:"smallMax"`  // Threshold for small
	MediumMax int `json:"mediumMax"` // Threshold for medium
}

// Scores holds derived composite scores for comparison.
type Scores struct {
	EfficiencyScore float64 `json:"efficiencyScore"` // 0-100
	QualityScore    float64 `json:"qualityScore"`    // 0-100
}

// computeTimeInState calculates the average days spent in each state
// across all work items using their state change history.
func computeTimeInState(workItems []WorkItem) []TimeInState {
	// Accumulate total days per state and count of items that transitioned through it
	stateDays := make(map[string]float64)
	stateCount := make(map[string]int)

	for _, wi := range workItems {
		if len(wi.History) == 0 {
			continue
		}

		// Walk through state changes and calculate time spent in each state
		for i, change := range wi.History {
			if change.ToState == "" {
				continue
			}

			// Duration in this state = time until the next state change (or until now/ChangedDate)
			var endTime time.Time
			if i+1 < len(wi.History) {
				endTime = wi.History[i+1].Date
			} else {
				endTime = wi.ChangedDate
			}

			if change.Date.IsZero() || endTime.IsZero() {
				continue
			}

			days := endTime.Sub(change.Date).Hours() / 24
			if days < 0 {
				continue
			}

			stateDays[change.ToState] += days
			stateCount[change.ToState]++
		}
	}

	var result []TimeInState
	for state, totalDays := range stateDays {
		count := stateCount[state]
		if count > 0 {
			result = append(result, TimeInState{
				State:   state,
				AvgDays: math.Round(totalDays/float64(count)*100) / 100,
			})
		}
	}

	return result
}

// computeThroughputTrend buckets completed items by week (Monday start).
// A work item is counted in the week its ReassignedDate falls in.
func computeThroughputTrend(workItems []WorkItem, from, to string) []ThroughputWeek {
	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil
	}
	toTime, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil
	}

	// Align fromTime to Monday
	for fromTime.Weekday() != time.Monday {
		fromTime = fromTime.AddDate(0, 0, -1)
	}

	// Build week buckets
	type weekBucket struct {
		start time.Time
		count int
	}
	var buckets []weekBucket
	for w := fromTime; !w.After(toTime); w = w.AddDate(0, 0, 7) {
		buckets = append(buckets, weekBucket{start: w})
	}

	// Place each completed item in a bucket
	for _, wi := range workItems {
		if wi.ReassignedDate == nil {
			continue
		}
		for i := range buckets {
			weekEnd := buckets[i].start.AddDate(0, 0, 7)
			if (wi.ReassignedDate.Equal(buckets[i].start) || wi.ReassignedDate.After(buckets[i].start)) &&
				wi.ReassignedDate.Before(weekEnd) {
				buckets[i].count++
				break
			}
		}
	}

	result := make([]ThroughputWeek, len(buckets))
	for i, b := range buckets {
		result[i] = ThroughputWeek{
			WeekStart: b.start.Format("2006-01-02"),
			Count:     b.count,
		}
	}
	return result
}

// computePRSizeBuckets categorizes PRs by files changed using configurable thresholds.
func computePRSizeBuckets(prDetails []PRDetail, filesPerPR map[int]int, smallMax, mediumMax int) PRSizeBuckets {
	buckets := PRSizeBuckets{SmallMax: smallMax, MediumMax: mediumMax}
	for _, pr := range prDetails {
		files := filesPerPR[pr.ID]
		switch {
		case files > mediumMax:
			buckets.Large++
		case files > smallMax:
			buckets.Medium++
		default:
			buckets.Small++
		}
	}
	return buckets
}

// computeScores calculates Efficiency and Quality scores for a report.
//
// Efficiency Score (0-100):
//   Measures how much a developer delivers relative to time spent.
//   Formula: normalized(resolved / avgResolutionDays) * priority weight boost.
//   Higher = more delivered faster on higher-priority work.
//
// Quality Score (0-100):
//   Measures how cleanly work is delivered.
//   Formula: (1 - reopenRate/100) * 70 + (1 - actionableCommentRate) * 30
//   Higher = fewer bouncebacks and fewer review iterations needed.
func computeScores(report *DeveloperReport) Scores {
	var scores Scores

	// --- Efficiency Score ---
	// Measures throughput over the reporting period, adjusted for item complexity (priority).
	// Base: items resolved per week in the reporting period.
	// A score of 100 = 5+ items/week on high-priority work.
	fromTime, err1 := time.Parse("2006-01-02", report.From)
	toTime, err2 := time.Parse("2006-01-02", report.To)
	if err1 == nil && err2 == nil && report.Resolved > 0 {
		weeks := toTime.Sub(fromTime).Hours() / 24 / 7
		if weeks < 1 {
			weeks = 1
		}
		itemsPerWeek := float64(report.Resolved) / weeks

		// Priority weight: boost for higher-priority work (up to 1.5x)
		priorityWeight := 1.0
		if report.TotalWorkedOn > 0 {
			highPriorityRatio := float64(report.Priority1Count+report.Priority2Count) / float64(report.TotalWorkedOn)
			priorityWeight = 1.0 + (highPriorityRatio * 0.5)
		}

		// Speed bonus: if avg resolution < 5 days, slight boost (up to 1.3x)
		speedWeight := 1.0
		if report.AvgResolutionDays > 0 && report.AvgResolutionDays < 5 {
			speedWeight = 1.0 + (5.0-report.AvgResolutionDays)/5.0*0.3
		}

		// Normalize: 5 items/week (with max weights) = 100
		raw := itemsPerWeek * priorityWeight * speedWeight
		scores.EfficiencyScore = math.Min(100, math.Round(raw/5.0*100*100)/100)
	} else if report.Resolved > 0 {
		// Fallback: partial credit based on raw count
		scores.EfficiencyScore = math.Min(100, float64(report.Resolved)*5)
	}

	// --- Quality Score ---
	// Component 1 (70%): Inverse reopen rate
	reopenComponent := (1.0 - report.ReopenRate/100.0) * 70.0

	// Component 2 (30%): Low actionable comment rate
	actionableComponent := 30.0 // default full marks if no PR data
	if report.PRMetrics != nil && report.PRMetrics.PRsMerged > 0 {
		commentRate := float64(report.PRMetrics.ActionableComments) / float64(report.PRMetrics.PRsMerged)
		// Normalize: 0 comments per PR = full marks, 5+ = 0 marks
		normalized := math.Max(0, 1.0-(commentRate/5.0))
		actionableComponent = normalized * 30.0
	}

	scores.QualityScore = math.Round((reopenComponent+actionableComponent)*100) / 100
	// Clamp
	scores.QualityScore = math.Max(0, math.Min(100, scores.QualityScore))

	return scores
}
