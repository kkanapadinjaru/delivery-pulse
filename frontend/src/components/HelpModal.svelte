<script>
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  const dispatch = createEventDispatcher();

  onMount(() => {
    document.body.style.overflow = 'hidden';
  });

  onDestroy(() => {
    document.body.style.overflow = '';
  });

  function close() {
    dispatch('close');
  }

  function handleBackdropClick(e) {
    if (e.target === e.currentTarget) close();
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') close();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="backdrop" on:click={handleBackdropClick} role="dialog" aria-modal="true" aria-label="Metrics Help">
  <div class="modal">
    <button class="close-btn" on:click={close} aria-label="Close">&times;</button>

    <div class="modal-body">
      <h1>Metrics Help</h1>

      <section>
        <h2>Work Item Metrics</h2>

        <div class="metric">
          <h3>Total Work Items <span class="badge volume">Volume</span></h3>
          <p>Count of work items (of the configured types) that were ever assigned to the developer and had activity within the selected date range.</p>
        </div>

        <div class="metric">
          <h3>Completed <span class="badge quality">Quality</span></h3>
          <p>Items where the developer handed off the work item to someone else (typically back to QA or a reviewer). Determined by the <strong>first</strong> time AssignedTo changed FROM the developer TO another person.</p>
        </div>

        <div class="metric">
          <h3>Avg Days to Complete <span class="badge quality">Quality</span></h3>
          <p>Average calendar days from when the developer was first assigned an item to when they first reassigned it away. Only the first assignment cycle counts.</p>
        </div>

        <div class="metric">
          <h3>Bounced Back <span class="badge quality">Quality</span></h3>
          <p>Items reassigned <strong>back to the developer</strong> after they handed them off. Indicates the reviewer/tester found the work incomplete. Based on assignment changes, not state transitions.</p>
        </div>

        <div class="metric">
          <h3>Reopen Rate <span class="badge quality">Quality</span></h3>
          <p>Percentage of items that bounced back. Thresholds:</p>
          <ul class="threshold-list">
            <li><span class="dot" style="background:#0f9d58"></span> <strong>Excellent</strong> — 0%</li>
            <li><span class="dot" style="background:#4caf50"></span> <strong>Good</strong> — 1–10%</li>
            <li><span class="dot" style="background:#ff9800"></span> <strong>Needs Improvement</strong> — 11–25%</li>
            <li><span class="dot" style="background:#f44336"></span> <strong>Concerning</strong> — &gt; 25%</li>
          </ul>
        </div>

        <div class="metric">
          <h3>Priority 1 / Priority 2 <span class="badge volume">Volume</span></h3>
          <p>Breakdown by priority, where priority is set on the item.</p>
        </div>
      </section>

      <section>
        <h2>Pull Request Metrics</h2>

        <div class="metric">
          <h3>PRs Merged <span class="badge volume">Volume</span></h3>
          <p>PRs created by the developer in the date range that were successfully completed (merged) across all repos.</p>
        </div>

        <div class="metric">
          <h3>Total Commits <span class="badge volume">Volume</span></h3>
          <p>Sum of all commits across all PRs in the date range.</p>
        </div>

        <div class="metric">
          <h3>Avg PR Cycle (days) <span class="badge quality">Quality</span></h3>
          <p>Average days from PR creation to merge. Shorter = faster review turnaround.</p>
        </div>

        <div class="metric">
          <h3>Files Changed <span class="badge volume">Volume</span></h3>
          <p>Total files modified across all merged PRs.</p>
        </div>

        <div class="metric">
          <h3>Actionable Comments <span class="badge quality">Quality</span></h3>
          <p>Reviewer comments that were followed by new code pushes, meaning the developer made changes to address feedback. Higher count = more rework needed before merge.</p>
        </div>

        <div class="metric">
          <h3>Rework PRs <span class="badge quality">Quality</span></h3>
          <p>Work items with multiple PRs to the <strong>same repository</strong>. Indicates the initial PR didn't fully address the requirement. Cross-repo PRs (e.g., frontend + backend) are expected and not flagged.</p>
        </div>
      </section>

      <section>
        <h2>Compare Mode — Relative Ranking</h2>
        <p>The ranking system scores each developer across weighted metrics. A lower total score = better overall performance.</p>
        <div class="metric">
          <h3>How it works</h3>
          <p>For each metric, developers are ranked (1 = best). The rank is multiplied by the metric's weight, then all weighted ranks are summed. The developer with the lowest total is ranked #1.</p>
        </div>
        <div class="metric">
          <h3>Metric Weights</h3>
          <ul class="threshold-list">
            <li><span class="dot" style="background:#206473"></span> <strong>Completed</strong> — 3× (highest priority: actual delivery)</li>
            <li><span class="dot" style="background:#206473"></span> <strong>Work Items</strong> — 2× (volume of work touched)</li>
            <li><span class="dot" style="background:#206473"></span> <strong>PRs Merged</strong> — 2× (code output)</li>
            <li><span class="dot" style="background:#45a6bd"></span> <strong>Reopen Rate</strong> — 1× (quality signal)</li>
            <li><span class="dot" style="background:#45a6bd"></span> <strong>Avg Days to Complete</strong> — 1× (speed)</li>
            <li><span class="dot" style="background:#45a6bd"></span> <strong>PR Cycle Time</strong> — 1× (review speed)</li>
            <li><span class="dot" style="background:#45a6bd"></span> <strong>Actionable Comments</strong> — 1× (review rework)</li>
          </ul>
          <p>This means volume/throughput accounts for 7 out of 11 total weight points, ensuring developers who deliver more work are ranked higher even if quality metrics are slightly worse.</p>
        </div>
      </section>

      <section>
        <h2>How This Helps Assess Delivery Quality</h2>
        <p><strong>Throughput:</strong> Total Work Items + PRs Merged show volume of work processed.</p>
        <p><strong>Completion quality:</strong> Bounced Back + Actionable Comments show if work is complete on first attempt.</p>
        <p><strong>Responsiveness:</strong> Avg Days to Complete and Avg PR Cycle show turnaround speed.</p>
        <p><strong>Review engagement:</strong> Actionable Comments show how much feedback translates to code changes.</p>
      </section>

      <section>
        <h2>Derived Insights (Single Report)</h2>

        <div class="metric">
          <h3>Time in State <span class="badge quality">Quality</span></h3>
          <p>Average days work items spend in each workflow state (e.g., New, Active, Resolved). Calculated from state change history. Highlights bottlenecks — if items sit in "Active" for 15 days on average, that's a signal.</p>
        </div>

        <div class="metric">
          <h3>Throughput Trend <span class="badge volume">Volume</span></h3>
          <p>Items completed per week over the report period, shown as a bar chart. A completed item is one the developer handed off (first reassignment). Weeks align to Monday. Helps spot velocity dips or ramp-ups.</p>
        </div>

        <div class="metric">
          <h3>PR Size Distribution <span class="badge quality">Quality</span></h3>
          <p>PRs categorized by files changed: <strong>Small</strong> (&lt;50 files), <strong>Medium</strong> (50–150), <strong>Large</strong> (&gt;150). Large PRs are harder to review thoroughly and correlate with missed defects. Aim for mostly small PRs.</p>
        </div>
      </section>

      <section>
        <h2>Composite Scores (Single Report &amp; Compare)</h2>

        <div class="metric">
          <h3>Efficiency Score <span class="badge volume">Volume</span></h3>
          <p>0–100 score measuring delivery throughput over the reporting period. Formula: items resolved per week, boosted by a priority weight (more P1/P2 work gives up to 1.5× boost) and a speed bonus (avg resolution under 5 days gives up to 1.3× boost). A score of 100 means 5+ items/week at max weights. Longer reporting periods give more meaningful scores.</p>
          <ul class="threshold-list">
            <li><span class="dot" style="background:#0f9d58"></span> <strong>Excellent</strong> — 80–100</li>
            <li><span class="dot" style="background:#4caf50"></span> <strong>Good</strong> — 60–79</li>
            <li><span class="dot" style="background:#ff9800"></span> <strong>Fair</strong> — 40–59</li>
            <li><span class="dot" style="background:#f44336"></span> <strong>Needs Attention</strong> — 0–39</li>
          </ul>
        </div>

        <div class="metric">
          <h3>Quality Score <span class="badge quality">Quality</span></h3>
          <p>0–100 score measuring how cleanly work is delivered. Two components:</p>
          <ul class="threshold-list">
            <li><span class="dot" style="background:#206473"></span> <strong>70%</strong> — Inverse reopen rate (fewer bouncebacks = higher score)</li>
            <li><span class="dot" style="background:#45a6bd"></span> <strong>30%</strong> — Low actionable comment rate (fewer reviewer-driven rework iterations = higher score)</li>
          </ul>
          <p>A score of 100 means zero bouncebacks and zero actionable review comments.</p>
          <ul class="threshold-list">
            <li><span class="dot" style="background:#0f9d58"></span> <strong>Excellent</strong> — 80–100</li>
            <li><span class="dot" style="background:#4caf50"></span> <strong>Good</strong> — 60–79</li>
            <li><span class="dot" style="background:#ff9800"></span> <strong>Fair</strong> — 40–59</li>
            <li><span class="dot" style="background:#f44336"></span> <strong>Needs Attention</strong> — 0–39</li>
          </ul>
        </div>
      </section>

      <section>
        <h2>Team Dashboard</h2>

        <div class="metric">
          <h3>Total Completed <span class="badge volume">Volume</span></h3>
          <p>Deduplicated count of all work items completed (handed off) across the entire team in the date range. An item assigned to multiple developers is counted once.</p>
        </div>

        <div class="metric">
          <h3>Work in Progress (WIP) <span class="badge quality">Quality</span></h3>
          <p>Items currently assigned to team members that are not yet Closed, Resolved, or Done — and haven't been reassigned away. High team WIP correlates with context switching and slower delivery.</p>
        </div>

        <div class="metric">
          <h3>Avg Cycle Time <span class="badge quality">Quality</span></h3>
          <p>Average calendar days from first assignment to first handoff, across all completed items for the team. This is the team's delivery rhythm — lower is faster.</p>
        </div>

        <div class="metric">
          <h3>Cycle Time Distribution <span class="badge quality">Quality</span></h3>
          <p>Histogram of cycle times in buckets (0–2 days, 3–5, 6–10, 11–20, 21–30, 30+). Shows whether most work ships quickly or if there's a long tail of items taking weeks. Useful for predicting delivery timelines.</p>
        </div>

        <div class="metric">
          <h3>WIP by Member <span class="badge quality">Quality</span></h3>
          <p>Per-person breakdown of in-progress items. Identifies individuals who may be overloaded with concurrent work.</p>
        </div>

        <div class="metric">
          <h3>Weekly Throughput <span class="badge volume">Volume</span></h3>
          <p>Same as the single-report throughput trend, but aggregated across the whole team. Shows team-level delivery cadence over time.</p>
        </div>
      </section>
    </div>
  </div>
</div>

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: flex-start;
    justify-content: center;
    z-index: 1000;
    padding: 2rem;
    overflow-y: auto;
  }

  .modal {
    background: white;
    border-radius: 12px;
    max-width: 720px;
    width: 100%;
    max-height: calc(100vh - 4rem);
    overflow-y: auto;
    position: relative;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    margin-top: 1rem;
  }

  .close-btn {
    position: sticky;
    top: 0;
    float: right;
    margin: 1rem 1rem 0 0;
    background: none;
    border: none;
    font-size: 1.8rem;
    color: #666;
    cursor: pointer;
    line-height: 1;
    z-index: 1;
  }

  .close-btn:hover {
    color: #206473;
  }

  .modal-body {
    padding: 2rem 2.5rem;
  }

  h1 {
    font-family: 'Manrope', 'Inter', system-ui, sans-serif;
    color: #206473;
    font-size: 1.4rem;
    margin-bottom: 1.5rem;
  }

  h2 {
    font-family: 'Manrope', 'Inter', system-ui, sans-serif;
    color: #206473;
    font-size: 1.1rem;
    margin-bottom: 0.75rem;
    padding-bottom: 0.4rem;
    border-bottom: 2px solid #e2e8f0;
  }

  section {
    margin-bottom: 1.5rem;
  }

  .metric {
    margin-bottom: 1rem;
    padding-left: 0.75rem;
    border-left: 3px solid #206473;
  }

  .metric h3 {
    font-size: 0.9rem;
    color: #2e3a3d;
    margin-bottom: 0.2rem;
  }

  .metric p {
    font-size: 0.85rem;
    color: #475569;
    line-height: 1.5;
    margin-bottom: 0.3rem;
  }

  .badge {
    display: inline-block;
    font-size: 0.65rem;
    padding: 0.1rem 0.4rem;
    border-radius: 3px;
    font-weight: 500;
    vertical-align: middle;
  }

  .badge.quality { background: #e8f4f7; color: #206473; }
  .badge.volume { background: #fff3e0; color: #e65100; }

  .threshold-list {
    list-style: none;
    padding: 0;
    margin: 0.3rem 0 0 0;
  }

  .threshold-list li {
    font-size: 0.82rem;
    color: #475569;
    padding: 0.15rem 0;
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  section:last-child p {
    font-size: 0.85rem;
    color: #475569;
    margin-bottom: 0.5rem;
    line-height: 1.6;
  }
</style>
