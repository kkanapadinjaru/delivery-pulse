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
        <h2>How This Helps Assess Delivery Quality</h2>
        <p><strong>Throughput:</strong> Total Work Items + PRs Merged show volume of work processed.</p>
        <p><strong>Completion quality:</strong> Bounced Back + Actionable Comments show if work is complete on first attempt.</p>
        <p><strong>Responsiveness:</strong> Avg Days to Complete and Avg PR Cycle show turnaround speed.</p>
        <p><strong>Review engagement:</strong> Actionable Comments show how much feedback translates to code changes.</p>
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
