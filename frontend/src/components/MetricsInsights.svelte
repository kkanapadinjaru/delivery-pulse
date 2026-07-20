<script>
  import { onMount } from 'svelte';
  export let report;

  // PR Size Buckets
  $: prSizeBuckets = report.prSizeBuckets || null;
  $: totalPRs = prSizeBuckets ? (prSizeBuckets.small + prSizeBuckets.medium + prSizeBuckets.large) : 0;

  // Scores
  $: scores = report.scores || null;

  function getScoreColor(score) {
    if (score >= 80) return '#0f9d58';
    if (score >= 60) return '#4caf50';
    if (score >= 40) return '#ff9800';
    return '#f44336';
  }

  function getScoreLabel(score) {
    if (score >= 80) return 'Excellent';
    if (score >= 60) return 'Good';
    if (score >= 40) return 'Fair';
    return 'Needs Attention';
  }
</script>

<section class="metrics-insights" aria-label="Metrics Insights">
  {#if scores}
  <div class="scores-row">
    <div class="score-card">
      <div class="score-ring" style="--score-color: {getScoreColor(scores.efficiencyScore)}; --score-pct: {scores.efficiencyScore}%">
        <span class="score-value">{scores.efficiencyScore.toFixed(0)}</span>
      </div>
      <div class="score-info">
        <span class="score-label">Efficiency</span>
        <span class="score-rating" style="color: {getScoreColor(scores.efficiencyScore)}">{getScoreLabel(scores.efficiencyScore)}</span>
      </div>
    </div>
    <div class="score-card">
      <div class="score-ring" style="--score-color: {getScoreColor(scores.qualityScore)}; --score-pct: {scores.qualityScore}%">
        <span class="score-value">{scores.qualityScore.toFixed(0)}</span>
      </div>
      <div class="score-info">
        <span class="score-label">Quality</span>
        <span class="score-rating" style="color: {getScoreColor(scores.qualityScore)}">{getScoreLabel(scores.qualityScore)}</span>
      </div>
    </div>
  </div>
  {/if}

  {#if prSizeBuckets && totalPRs > 0}
  <div class="insight-section">
    <h3>PR Size Distribution</h3>
    <p class="section-help">PRs categorized by files changed. Small (&le;{prSizeBuckets.smallMax}), Medium ({prSizeBuckets.smallMax + 1}-{prSizeBuckets.mediumMax}), Large (&gt;{prSizeBuckets.mediumMax}).</p>
    <div class="size-buckets">
      <div class="bucket">
        <div class="bucket-bar-wrapper">
          <div class="bucket-bar small" style="height: {(prSizeBuckets.small / totalPRs) * 100}%"></div>
        </div>
        <span class="bucket-count">{prSizeBuckets.small}</span>
        <span class="bucket-label">Small</span>
      </div>
      <div class="bucket">
        <div class="bucket-bar-wrapper">
          <div class="bucket-bar medium" style="height: {(prSizeBuckets.medium / totalPRs) * 100}%"></div>
        </div>
        <span class="bucket-count">{prSizeBuckets.medium}</span>
        <span class="bucket-label">Medium</span>
      </div>
      <div class="bucket">
        <div class="bucket-bar-wrapper">
          <div class="bucket-bar large" style="height: {(prSizeBuckets.large / totalPRs) * 100}%"></div>
        </div>
        <span class="bucket-count">{prSizeBuckets.large}</span>
        <span class="bucket-label">Large</span>
      </div>
    </div>
  </div>
  {/if}
</section>

<style>
  .metrics-insights {
    background: white;
    border-radius: 10px;
    padding: 1.5rem 2rem;
    margin-bottom: 1.5rem;
    box-shadow: 0 1px 4px rgba(46, 58, 61, 0.08), 0 4px 16px rgba(46, 58, 61, 0.06);
  }

  .scores-row {
    display: flex;
    gap: 2rem;
    margin-bottom: 1.5rem;
    padding-bottom: 1.5rem;
    border-bottom: 1px solid #e2e8f0;
  }

  .score-card {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .score-ring {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background: conic-gradient(var(--score-color) var(--score-pct), #e8e8e8 var(--score-pct));
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
  }

  .score-ring::before {
    content: '';
    position: absolute;
    width: 46px;
    height: 46px;
    border-radius: 50%;
    background: white;
  }

  .score-value {
    position: relative;
    z-index: 1;
    font-size: 1.1rem;
    font-weight: 700;
    color: #333;
  }

  .score-info {
    display: flex;
    flex-direction: column;
  }

  .score-label {
    font-size: 0.85rem;
    font-weight: 600;
    color: #333;
  }

  .score-rating {
    font-size: 0.75rem;
    font-weight: 600;
  }

  .insight-section {
    margin-bottom: 1.5rem;
  }

  .insight-section:last-child {
    margin-bottom: 0;
  }

  h3 {
    font-size: 0.95rem;
    color: #206473;
    margin-bottom: 0.25rem;
    font-weight: 600;
  }

  .section-help {
    font-size: 0.78rem;
    color: #888;
    margin-bottom: 0.75rem;
  }

  /* PR Size Buckets */
  .size-buckets {
    display: flex;
    gap: 2rem;
    align-items: flex-end;
    height: 130px;
    justify-content: center;
    padding-top: 1rem;
  }

  .bucket {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.3rem;
  }

  .bucket-bar-wrapper {
    width: 40px;
    height: 80px;
    display: flex;
    align-items: flex-end;
  }

  .bucket-bar {
    width: 100%;
    border-radius: 4px 4px 0 0;
    min-height: 4px;
    transition: height 0.3s ease;
  }

  .bucket-bar.small { background: #0f9d58; }
  .bucket-bar.medium { background: #ff9800; }
  .bucket-bar.large { background: #f44336; }

  .bucket-count {
    font-size: 1rem;
    font-weight: 700;
    color: #333;
  }

  .bucket-label {
    font-size: 0.72rem;
    color: #666;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }
</style>
