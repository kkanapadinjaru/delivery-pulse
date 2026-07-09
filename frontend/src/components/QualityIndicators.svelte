<script>
  export let report;
  export let prDetails = [];

  $: reopenRate = report.reopenRate?.toFixed(1) || '0';
  $: qualityRating = getQualityRating(report.reopenRate || 0);
  $: hasSimilarBugs = report.similarBugs && report.similarBugs.length > 0;

  // Compute rework PRs: work items with multiple PRs to the same repo
  $: reworkData = (() => {
    const prByWorkItem = {};
    for (const pr of prDetails) {
      if (pr.linkedWorkItemIds) {
        for (const wiId of pr.linkedWorkItemIds) {
          if (!prByWorkItem[wiId]) prByWorkItem[wiId] = [];
          prByWorkItem[wiId].push(pr);
        }
      }
    }
    let reworkItemCount = 0;
    let totalReworkPRs = 0;
    for (const [, prs] of Object.entries(prByWorkItem)) {
      const byRepo = {};
      for (const pr of prs) {
        const repo = pr.repository || 'Unknown';
        byRepo[repo] = (byRepo[repo] || 0) + 1;
      }
      const hasRework = Object.values(byRepo).some(count => count > 1);
      if (hasRework) {
        reworkItemCount++;
        totalReworkPRs += Object.values(byRepo).filter(c => c > 1).reduce((sum, c) => sum + c, 0);
      }
    }
    return { itemCount: reworkItemCount, prCount: totalReworkPRs };
  })();

  function getQualityRating(rate) {
    if (rate === 0) return { label: 'Excellent', color: '#0f9d58', icon: 'check-circle', tooltip: '0% — No items bounced back' };
    if (rate <= 10) return { label: 'Good', color: '#4caf50', icon: 'check', tooltip: '1–10% — Low bounce-back rate' };
    if (rate <= 25) return { label: 'Needs Improvement', color: '#ff9800', icon: 'alert', tooltip: '11–25% — Moderate bounce-back rate' };
    return { label: 'Concerning', color: '#f44336', icon: 'warning', tooltip: '> 25% — High bounce-back rate' };
  }
</script>

<section class="quality" aria-label="Delivery Quality Indicators">
  <h2>Delivery Quality</h2>

  <div class="quality-grid">
    <div class="quality-card reopen">
      <div class="quality-header">
        <h3>Reopen Rate</h3>
        <span class="badge-wrapper">
          <span class="badge" style="background: {qualityRating.color}">
            {qualityRating.label}
          </span>
          <span class="info-trigger" aria-label="Threshold info">
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="8" cy="8" r="7" stroke="#666" stroke-width="1.5" fill="none"/>
              <path d="M8 7v4" stroke="#666" stroke-width="1.5" stroke-linecap="round"/>
              <circle cx="8" cy="5" r="0.75" fill="#666"/>
            </svg>
            <span class="tooltip-popup">{qualityRating.tooltip}</span>
          </span>
        </span>
      </div>
      <div class="quality-stat">
        <span class="big-num">{reopenRate}%</span>
        <span class="detail">
          {report.reopened} of {report.totalWorkedOn} items bounced back
        </span>
      </div>
      <div class="quality-bar">
        <div
          class="quality-fill"
          style="width: {Math.min(reopenRate, 100)}%; background: {qualityRating.color}"
        ></div>
      </div>
    </div>

    <div class="quality-card similar">
      <div class="quality-header">
        <h3>Similar / Recurring Bugs</h3>
      </div>
      {#if hasSimilarBugs}
        <p class="warning-text">
          Found {report.similarBugs.length} potentially related bug pair(s)
          that may indicate incomplete fixes.
        </p>
        <ul class="similar-list">
          {#each report.similarBugs as bug}
            <li>
              <span class="similar-pair">
                <strong>#{bug.originalId}</strong> "{bug.originalTitle}"
                <span class="arrow">&#8594;</span>
                <strong>#{bug.newId}</strong> "{bug.newTitle}"
              </span>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="ok-text">No similar/recurring bugs detected. Fixes appear thorough.</p>
      {/if}
    </div>

    <div class="quality-card rework">
      <div class="quality-header">
        <h3>Rework PRs</h3>
      </div>
      {#if reworkData.itemCount > 0}
        <div class="quality-stat">
          <span class="big-num">{reworkData.itemCount}</span>
          <span class="detail">
            work item{reworkData.itemCount !== 1 ? 's' : ''} with multiple PRs to the same repo ({reworkData.prCount} PRs total)
          </span>
        </div>
        <p class="warning-text-sm">
          May indicate incomplete initial submissions requiring follow-up fixes.
        </p>
      {:else}
        <p class="ok-text">No rework detected. Each work item has at most one PR per repository.</p>
      {/if}
    </div>
  </div>
</section>

<style>
  .quality {
    background: white;
    border-radius: 10px;
    padding: 1.5rem 2rem;
    margin-bottom: 1.5rem;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  }

  h2 {
    font-size: 1.1rem;
    margin-bottom: 1rem;
    color: #333;
  }

  .quality-grid {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 1.5rem;
  }

  @media (max-width: 960px) {
    .quality-grid {
      grid-template-columns: 1fr 1fr;
    }
  }

  @media (max-width: 768px) {
    .quality-grid {
      grid-template-columns: 1fr;
    }
  }

  .quality-card {
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 1.25rem;
  }

  .quality-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.75rem;
  }

  h3 {
    font-size: 0.95rem;
    color: #444;
  }

  .badge {
    font-size: 0.75rem;
    padding: 0.2rem 0.6rem;
    border-radius: 12px;
    color: white;
    font-weight: 600;
  }

  .badge-wrapper {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
  }

  .info-trigger {
    position: relative;
    display: inline-flex;
    align-items: center;
    cursor: help;
  }

  .info-trigger svg {
    display: block;
  }

  .tooltip-popup {
    display: none;
    position: absolute;
    bottom: calc(100% + 6px);
    right: 0;
    background: #1a1a2e;
    color: white;
    font-size: 0.72rem;
    font-weight: 500;
    padding: 0.4rem 0.65rem;
    border-radius: 5px;
    white-space: nowrap;
    z-index: 50;
    box-shadow: 0 2px 8px rgba(0,0,0,0.2);
  }

  .tooltip-popup::after {
    content: '';
    position: absolute;
    top: 100%;
    right: 4px;
    border: 5px solid transparent;
    border-top-color: #1a1a2e;
  }

  .info-trigger:hover .tooltip-popup {
    display: block;
  }

  .quality-stat {
    margin-bottom: 0.75rem;
  }

  .big-num {
    font-size: 2rem;
    font-weight: 700;
    color: #1a1a2e;
  }

  .detail {
    display: block;
    font-size: 0.8rem;
    color: #666;
    margin-top: 0.25rem;
  }

  .quality-bar {
    height: 6px;
    background: #e8e8e8;
    border-radius: 3px;
    overflow: hidden;
  }

  .quality-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.5s ease;
  }

  .warning-text {
    color: #e65100;
    font-size: 0.9rem;
    margin-bottom: 0.75rem;
  }

  .warning-text-sm {
    color: #e65100;
    font-size: 0.8rem;
    margin-top: 0.5rem;
  }

  .ok-text {
    color: #2e7d32;
    font-size: 0.9rem;
  }

  .similar-list {
    list-style: none;
    padding: 0;
  }

  .similar-list li {
    padding: 0.5rem 0;
    border-bottom: 1px solid #f0f0f0;
    font-size: 0.85rem;
  }

  .similar-list li:last-child {
    border-bottom: none;
  }

  .arrow {
    color: #999;
    margin: 0 0.5rem;
  }
</style>
