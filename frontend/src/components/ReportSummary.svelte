<script>
  export let report;

  // Compute work item type counts from report.workItems
  $: typeCounts = (report.workItems || []).reduce((acc, wi) => {
    const type = wi.workItemType || 'Unknown';
    acc[type] = (acc[type] || 0) + 1;
    return acc;
  }, {});

  $: typeEntries = Object.entries(typeCounts).sort((a, b) => b[1] - a[1]);

  // Compute type + priority breakdown
  $: typePriorityData = (report.workItems || []).reduce((acc, wi) => {
    const type = wi.workItemType || 'Unknown';
    const priority = wi.priority || 0;
    const key = `${type}|||${priority}`;
    acc[key] = (acc[key] || 0) + 1;
    return acc;
  }, {});

  // Group into rows: type, priority, count — sorted by type then priority
  $: typePriorityRows = Object.entries(typePriorityData)
    .map(([key, count]) => {
      const [type, priority] = key.split('|||');
      return { type, priority: parseInt(priority), count };
    })
    .sort((a, b) => a.type.localeCompare(b.type) || a.priority - b.priority);
</script>

<section class="summary" aria-label="Report Summary">
  <h2>Performance Summary: {report.developer}</h2>
  <p class="date-range">{report.from} to {report.to}</p>

  <div class="cards">
    <div class="card">
      <span class="card-value">{report.totalWorkedOn}</span>
      <span class="card-label">Total Work Items</span>
    </div>
    <div class="card">
      <span class="card-value">{report.resolved}</span>
      <span class="card-label">Completed</span>
    </div>
    <div class="card priority">
      <span class="card-value">{report.priority1Count}</span>
      <span class="card-label">Priority 1</span>
    </div>
    <div class="card priority">
      <span class="card-value">{report.priority2Count}</span>
      <span class="card-label">Priority 2</span>
    </div>
    <div class="card">
      <span class="card-value">{report.avgResolutionDays?.toFixed(1) || 'N/A'}</span>
      <span class="card-label">Avg Days to Complete</span>
    </div>
    <div class="card quality">
      <span class="card-value">{report.reopened}</span>
      <span class="card-label">Bounced Back</span>
    </div>
  </div>

  {#if typePriorityRows.length > 0}
  <h3 class="section-title">By Work Item Type</h3>
  <div class="type-table-wrapper">
    <table class="type-table">
      <thead>
        <tr>
          <th>Type</th>
          <th>Priority</th>
          <th>Count</th>
        </tr>
      </thead>
      <tbody>
        {#each typePriorityRows as row}
          <tr>
            <td class="type-cell">{row.type}</td>
            <td><span class="priority-badge p{row.priority}">P{row.priority}</span></td>
            <td class="count-cell">{row.count}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  {/if}

  {#if report.prMetrics}
  <h3 class="section-title">Pull Request Activity</h3>
  <div class="cards">
    <div class="card pr">
      <span class="card-value">{report.prMetrics.prsMerged}</span>
      <span class="card-label">PRs Merged</span>
    </div>
    <div class="card pr">
      <span class="card-value">{report.prMetrics.totalCommits}</span>
      <span class="card-label">Total Commits</span>
    </div>
    <div class="card pr">
      <span class="card-value">{report.prMetrics.avgPRCycleDays?.toFixed(1) || 'N/A'}</span>
      <span class="card-label">Avg PR Cycle (days)</span>
    </div>
    <div class="card pr">
      <span class="card-value">{report.prMetrics.filesChanged}</span>
      <span class="card-label">Files Changed</span>
    </div>
    <div class="card pr">
      <span class="card-value">{report.prMetrics.actionableComments}</span>
      <span class="card-label">Actionable Comments</span>
    </div>
  </div>
  {/if}
</section>

<style>
  .summary {
    background: white;
    border-radius: 10px;
    padding: 1.5rem 2rem;
    margin-bottom: 1.5rem;
    box-shadow: 0 1px 4px rgba(46, 58, 61, 0.08), 0 4px 16px rgba(46, 58, 61, 0.06);
  }

  h2 {
    font-family: 'Manrope', 'Inter', system-ui, sans-serif;
    font-size: 1.2rem;
    color: #206473;
    margin-bottom: 0.25rem;
  }

  .date-range {
    font-size: 0.85rem;
    color: #666;
    margin-bottom: 1.25rem;
  }

  .cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 1rem;
  }

  .card {
    background: #f8fafb;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 1rem;
    text-align: center;
  }

  .card-value {
    display: block;
    font-size: 1.85rem;
    font-weight: 700;
    color: #206473;
  }

  .card-label {
    font-size: 0.75rem;
    color: #666;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-top: 0.25rem;
    display: block;
  }

  .card.priority .card-value {
    color: #d95a32;
  }

  .card.quality .card-value {
    color: #c62828;
  }

  .card.pr .card-value {
    color: #185364;
  }

  .type-table-wrapper {
    overflow-x: auto;
  }

  .type-table {
    width: 100%;
    max-width: 400px;
    border-collapse: collapse;
    font-size: 0.85rem;
  }

  .type-table thead th {
    text-align: left;
    padding: 0.4rem 0.75rem;
    border-bottom: 2px solid #e2e8f0;
    font-weight: 600;
    color: #206473;
    font-size: 0.78rem;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .type-table tbody td {
    padding: 0.4rem 0.75rem;
    border-bottom: 1px solid #f0f0f0;
  }

  .type-table tbody tr:last-child td {
    border-bottom: none;
  }

  .type-cell {
    font-weight: 500;
    color: #333;
  }

  .count-cell {
    font-weight: 600;
    color: #206473;
  }

  .priority-badge {
    display: inline-block;
    font-size: 0.72rem;
    font-weight: 600;
    padding: 0.1rem 0.4rem;
    border-radius: 3px;
  }

  .priority-badge.p1 {
    background: #fde8e8;
    color: #c62828;
  }

  .priority-badge.p2 {
    background: #fff3e0;
    color: #e65100;
  }

  .section-title {
    font-size: 0.95rem;
    color: #206473;
    margin-top: 1.5rem;
    margin-bottom: 0.75rem;
    padding-top: 1rem;
    border-top: 1px solid #e2e8f0;
    font-weight: 600;
  }
</style>
