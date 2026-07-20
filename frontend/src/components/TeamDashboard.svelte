<script>
  import { onMount } from 'svelte';

  let from = '';
  let to = '';
  let dashboard = null;
  let loading = false;
  let error = '';
  let activePreset = 'Last 30d';

  // Date presets
  const now = new Date();
  const currentYear = now.getFullYear();
  const thirtyDaysAgo = new Date(now);
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
  const presets = [
    { label: 'Q1', from: `${currentYear}-01-01`, to: `${currentYear}-03-31` },
    { label: 'Q2', from: `${currentYear}-04-01`, to: `${currentYear}-06-30` },
    { label: 'Q3', from: `${currentYear}-07-01`, to: `${currentYear}-09-30` },
    { label: 'Q4', from: `${currentYear}-10-01`, to: `${currentYear}-12-31` },
    { label: 'HY1', from: `${currentYear}-01-01`, to: `${currentYear}-06-30` },
    { label: 'HY2', from: `${currentYear}-07-01`, to: `${currentYear}-12-31` },
    { label: 'YTD', from: `${currentYear}-01-01`, to: now.toISOString().split('T')[0] },
    { label: 'Last 30d', from: thirtyDaysAgo.toISOString().split('T')[0], to: now.toISOString().split('T')[0] },
  ];

  // Default to last 30 days
  from = thirtyDaysAgo.toISOString().split('T')[0];
  to = now.toISOString().split('T')[0];

  function applyPreset(preset) {
    from = preset.from;
    to = preset.to;
    activePreset = preset.label;
    loadDashboard();
  }

  function handleDateChange() {
    activePreset = '';
  }

  async function loadDashboard() {
    loading = true;
    error = '';
    dashboard = null;

    try {
      const params = new URLSearchParams({ from, to });
      const resp = await fetch(`/api/team-dashboard?${params}`);
      if (!resp.ok) {
        const data = await resp.json().catch(() => ({}));
        throw new Error(data.error || `Request failed (${resp.status})`);
      }
      dashboard = await resp.json();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  onMount(loadDashboard);

  $: maxThroughput = dashboard ? Math.max(...(dashboard.throughputTrend || []).map(w => w.count), 1) : 1;
  $: maxWIP = dashboard ? Math.max(...(dashboard.wipByMember || []).map(m => m.count), 1) : 1;
  $: cycleDistribution = dashboard?.cycleTimeDistribution || [];
  $: maxCycleBucket = Math.max(...cycleDistribution.map(b => b.count), 1);

  function shortName(email) {
    const local = email.split('@')[0] || email;
    return local.split('.').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ');
  }
</script>

<section class="team-dashboard" aria-label="Team Dashboard">
  <div class="dashboard-header">
    <h2>Team Dashboard</h2>
    <div class="date-controls">
      <div class="presets">
        {#each presets as preset}
          <button
            type="button"
            class="preset-btn"
            class:active={activePreset === preset.label}
            on:click={() => applyPreset(preset)}
          >{preset.label}</button>
        {/each}
      </div>
      <input type="date" bind:value={from} on:change={handleDateChange} />
      <span class="date-sep">to</span>
      <input type="date" bind:value={to} on:change={handleDateChange} />
      <button class="refresh-btn" on:click={loadDashboard} disabled={loading}>
        {loading ? 'Loading...' : 'Refresh'}
      </button>
    </div>
  </div>

  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>Loading team metrics...</p>
    </div>
  {/if}

  {#if error}
    <div class="error-state" role="alert">
      <strong>Error:</strong> {error}
    </div>
  {/if}

  {#if dashboard}
    <div class="summary-cards">
      <div class="card">
        <span class="card-value">{dashboard.totalMembers}</span>
        <span class="card-label">Team Members</span>
      </div>
      <div class="card">
        <span class="card-value">{dashboard.totalCompleted}</span>
        <span class="card-label">Completed</span>
      </div>
      <div class="card highlight">
        <span class="card-value">{dashboard.totalWip}</span>
        <span class="card-label">Work in Progress</span>
      </div>
      <div class="card">
        <span class="card-value">{dashboard.avgCycleTimeDays?.toFixed(1) || 'N/A'}</span>
        <span class="card-label">Avg Cycle Time (days)</span>
      </div>
    </div>

    <!-- Throughput Trend -->
    {#if dashboard.throughputTrend && dashboard.throughputTrend.length > 0}
    <div class="chart-section">
      <h3>Weekly Throughput</h3>
      <p class="chart-help">Items completed per week across the team.</p>
      <div class="bar-chart">
        {#each dashboard.throughputTrend as week}
          <div class="bar-col" title="{week.weekStart}: {week.count} items">
            <div class="bar-fill" style="height: {(week.count / maxThroughput) * 100}%">
              {#if week.count > 0}<span class="bar-value">{week.count}</span>{/if}
            </div>
            <span class="bar-label">{week.weekStart.slice(5)}</span>
          </div>
        {/each}
      </div>
    </div>
    {/if}

    <!-- Cycle Time Distribution -->
    {#if cycleDistribution.length > 0}
    <div class="chart-section">
      <h3>Cycle Time Distribution</h3>
      <p class="chart-help">How long items take from assignment to handoff. Helps predict delivery timelines.</p>
      <div class="histogram">
        {#each cycleDistribution as bucket}
          <div class="hist-col">
            <div class="hist-bar" style="height: {(bucket.count / maxCycleBucket) * 100}%">
              {#if bucket.count > 0}<span class="hist-value">{bucket.count}</span>{/if}
            </div>
            <span class="hist-label">{bucket.label}</span>
          </div>
        {/each}
      </div>
    </div>
    {/if}

    <!-- WIP by Member -->
    {#if dashboard.wipByMember && dashboard.wipByMember.filter(m => m.count > 0).length > 0}
    <div class="chart-section">
      <h3>WIP by Team Member</h3>
      <p class="chart-help">Currently in-progress items per person. High WIP often correlates with context switching.</p>
      <div class="wip-list">
        {#each dashboard.wipByMember.filter(m => m.count > 0).sort((a, b) => b.count - a.count) as member}
          <div class="wip-row">
            <span class="wip-name">{shortName(member.developer)}</span>
            <div class="wip-bar-track">
              <div class="wip-bar-fill" style="width: {(member.count / maxWIP) * 100}%"></div>
            </div>
            <span class="wip-count">{member.count}</span>
          </div>
        {/each}
      </div>
    </div>
    {/if}
  {/if}
</section>

<style>
  .team-dashboard {
    background: white;
    border-radius: 10px;
    padding: 1.5rem 2rem;
    box-shadow: 0 1px 4px rgba(46, 58, 61, 0.08), 0 4px 16px rgba(46, 58, 61, 0.06);
  }

  .dashboard-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
    flex-wrap: wrap;
    gap: 1rem;
  }

  h2 {
    font-family: 'Manrope', 'Inter', system-ui, sans-serif;
    font-size: 1.3rem;
    color: #206473;
    margin: 0;
  }

  .date-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .presets {
    display: flex;
    gap: 0.25rem;
    margin-right: 0.5rem;
  }

  .preset-btn {
    padding: 0.3rem 0.6rem;
    border: 1px solid #d0d5dd;
    border-radius: 4px;
    background: white;
    font-size: 0.72rem;
    font-weight: 500;
    color: #555;
    cursor: pointer;
    transition: all 0.15s;
  }

  .preset-btn:hover {
    border-color: #206473;
    color: #206473;
  }

  .preset-btn.active {
    background: #206473;
    border-color: #206473;
    color: white;
  }

  .date-controls input {
    padding: 0.4rem 0.6rem;
    border: 1px solid #d0d5dd;
    border-radius: 6px;
    font-size: 0.85rem;
  }

  .date-sep {
    font-size: 0.8rem;
    color: #666;
  }

  .refresh-btn {
    padding: 0.4rem 1rem;
    background: #206473;
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
  }

  .refresh-btn:hover:not(:disabled) { background: #185364; }
  .refresh-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 3rem;
    color: #555;
  }

  .spinner {
    width: 36px;
    height: 36px;
    border: 4px solid #e0e0e0;
    border-top-color: #206473;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin-bottom: 1rem;
  }

  @keyframes spin { to { transform: rotate(360deg); } }

  .error-state {
    background: #fde8e8;
    border: 1px solid #f5aca6;
    border-radius: 8px;
    padding: 1rem;
    color: #9b1c1c;
    margin-bottom: 1rem;
  }

  .summary-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }

  .card {
    background: #f8fafb;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 1.25rem 1rem;
    text-align: center;
  }

  .card.highlight {
    border-color: #ff9800;
    background: #fff8f0;
  }

  .card-value {
    display: block;
    font-size: 2rem;
    font-weight: 700;
    color: #206473;
  }

  .card.highlight .card-value { color: #e65100; }

  .card-label {
    font-size: 0.75rem;
    color: #666;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-top: 0.25rem;
    display: block;
  }

  .chart-section {
    margin-bottom: 2rem;
  }

  h3 {
    font-size: 0.95rem;
    color: #206473;
    margin-bottom: 0.25rem;
    font-weight: 600;
  }

  .chart-help {
    font-size: 0.78rem;
    color: #888;
    margin-bottom: 0.75rem;
  }

  /* Throughput Bar Chart */
  .bar-chart {
    display: flex;
    align-items: flex-end;
    gap: 4px;
    height: 140px;
    padding: 0.5rem 0;
  }

  .bar-col {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    height: 100%;
    justify-content: flex-end;
  }

  .bar-fill {
    width: 100%;
    max-width: 36px;
    background: #206473;
    border-radius: 3px 3px 0 0;
    min-height: 2px;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    position: relative;
  }

  .bar-value {
    position: absolute;
    top: -18px;
    font-size: 0.68rem;
    font-weight: 600;
    color: #206473;
  }

  .bar-label {
    font-size: 0.62rem;
    color: #888;
    margin-top: 4px;
  }

  /* Histogram */
  .histogram {
    display: flex;
    align-items: flex-end;
    gap: 8px;
    height: 120px;
  }

  .hist-col {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    height: 100%;
    justify-content: flex-end;
  }

  .hist-bar {
    width: 100%;
    max-width: 50px;
    background: #45a6bd;
    border-radius: 3px 3px 0 0;
    min-height: 2px;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    position: relative;
  }

  .hist-value {
    position: absolute;
    top: -18px;
    font-size: 0.7rem;
    font-weight: 600;
    color: #45a6bd;
  }

  .hist-label {
    font-size: 0.68rem;
    color: #666;
    margin-top: 6px;
    text-align: center;
  }

  /* WIP by Member */
  .wip-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .wip-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .wip-name {
    font-size: 0.8rem;
    font-weight: 500;
    color: #444;
    min-width: 140px;
  }

  .wip-bar-track {
    flex: 1;
    height: 10px;
    background: #f0f0f0;
    border-radius: 5px;
    overflow: hidden;
  }

  .wip-bar-fill {
    height: 100%;
    background: #e65100;
    border-radius: 5px;
    transition: width 0.3s ease;
  }

  .wip-count {
    font-size: 0.8rem;
    font-weight: 700;
    color: #e65100;
    min-width: 24px;
    text-align: right;
  }
</style>
