<script>
  import { onMount, onDestroy } from 'svelte';
  import Chart from 'chart.js/auto';

  export let reports = []; // Array of report objects

  // Set Chart.js global font to Inter
  Chart.defaults.font.family = "'Inter', system-ui, -apple-system, sans-serif";

  let bugChartCanvas;
  let prChartCanvas;
  let qualityChartCanvas;
  let bugChart;
  let prChart;
  let qualityChart;

  const colors = ['#206473', '#d95a32', '#45a6bd'];
  const bgColors = ['rgba(32,100,115,0.7)', 'rgba(217,90,50,0.7)', 'rgba(69,166,189,0.7)'];

  $: labels = reports.map(r => shortName(r.developer));

  // Collect all unique work item types across reports
  $: allTypes = [...new Set(reports.flatMap(r => (r.workItems || []).map(wi => wi.workItemType || 'Unknown')))].sort();

  // For each type, get the count per report
  function typeCountFor(report, type) {
    return (report.workItems || []).filter(wi => (wi.workItemType || 'Unknown') === type).length;
  }

  function shortName(email) {
    // Show first.last from email
    const local = email.split('@')[0] || email;
    return local.split('.').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ');
  }

  function getRateColor(rate) {
    if (rate === 0) return '#0f9d58';
    if (rate <= 10) return '#4caf50';
    if (rate <= 25) return '#ff9800';
    return '#f44336';
  }

  // Compute relative rankings across key metrics
  // Higher is better: totalWorkedOn, resolved, prsMerged, totalCommits
  // Lower is better: reopenRate, avgResolutionDays, avgPRCycleDays, actionableComments
  $: rankings = (() => {
    if (reports.length < 2) return [];

    const metrics = [
      { key: 'totalWorkedOn', label: 'Work Items', higher: true },
      { key: 'resolved', label: 'Completed', higher: true },
      { key: 'reopenRate', label: 'Reopen Rate', higher: false },
      { key: 'avgResolutionDays', label: 'Avg Days', higher: false },
      { key: 'prsMerged', label: 'PRs Merged', higher: true, nested: 'prMetrics' },
      { key: 'avgPRCycleDays', label: 'PR Cycle', higher: false, nested: 'prMetrics' },
      { key: 'actionableComments', label: 'Act. Comments', higher: false, nested: 'prMetrics' },
    ];

    // For each metric, rank developers (1 = best)
    const scores = reports.map((r, i) => ({
      index: i,
      name: shortName(r.developer),
      totalScore: 0,
      breakdown: [],
    }));

    for (const metric of metrics) {
      const values = reports.map(r => {
        const val = metric.nested ? r[metric.nested]?.[metric.key] : r[metric.key];
        return val ?? (metric.higher ? 0 : 999);
      });

      // Create ranked indices
      const indexed = values.map((v, i) => ({ v, i }));
      indexed.sort((a, b) => metric.higher ? b.v - a.v : a.v - b.v);

      // Assign ranks (handle ties)
      let currentRank = 1;
      for (let j = 0; j < indexed.length; j++) {
        if (j > 0 && indexed[j].v !== indexed[j - 1].v) {
          currentRank = j + 1;
        }
        const devIdx = indexed[j].i;
        scores[devIdx].totalScore += currentRank;
        scores[devIdx].breakdown.push({ metric: metric.label, rank: currentRank });
      }
    }

    // Sort by total score (lower = better)
    return scores.sort((a, b) => a.totalScore - b.totalScore)
      .map(s => ({ name: s.name, score: s.totalScore, breakdown: s.breakdown }));
  })();

  function buildCharts() {
    destroyCharts();

    if (!reports.length || !bugChartCanvas) return;

    const devLabels = labels;

    // Bug Fix metrics chart
    bugChart = new Chart(bugChartCanvas, {
      type: 'bar',
      data: {
        labels: devLabels,
        datasets: [
          {
            label: 'Work Items',
            data: reports.map(r => r.totalWorkedOn),
            backgroundColor: bgColors[0],
            borderColor: colors[0],
            borderWidth: 1,
          },
          {
            label: 'Completed',
            data: reports.map(r => r.resolved),
            backgroundColor: bgColors[1],
            borderColor: colors[1],
            borderWidth: 1,
          },
          {
            label: 'Bounced Back',
            data: reports.map(r => r.reopened),
            backgroundColor: bgColors[2],
            borderColor: colors[2],
            borderWidth: 1,
          },
        ],
      },
      options: chartOptions('Work Item Volume'),
    });

    // PR metrics chart
    const hasPR = reports.some(r => r.prMetrics);
    if (hasPR) {
      prChart = new Chart(prChartCanvas, {
        type: 'bar',
        data: {
          labels: devLabels,
          datasets: [
            {
              label: 'PRs Merged',
              data: reports.map(r => r.prMetrics?.prsMerged || 0),
              backgroundColor: bgColors[0],
              borderColor: colors[0],
              borderWidth: 1,
            },
            {
              label: 'Total Commits',
              data: reports.map(r => r.prMetrics?.totalCommits || 0),
              backgroundColor: bgColors[1],
              borderColor: colors[1],
              borderWidth: 1,
            },
          ],
        },
        options: chartOptions('Pull Request Volume'),
      });
    }

    // Quality metrics chart (days/rates)
    qualityChart = new Chart(qualityChartCanvas, {
      type: 'bar',
      data: {
        labels: devLabels,
        datasets: [
          {
            label: 'Avg Days to Complete',
            data: reports.map(r => r.avgResolutionDays ? +r.avgResolutionDays.toFixed(1) : 0),
            backgroundColor: bgColors[0],
            borderColor: colors[0],
            borderWidth: 1,
          },
          {
            label: 'Avg PR Cycle (days)',
            data: reports.map(r => r.prMetrics?.avgPRCycleDays ? +r.prMetrics.avgPRCycleDays.toFixed(1) : 0),
            backgroundColor: bgColors[1],
            borderColor: colors[1],
            borderWidth: 1,
          },
          {
            label: 'Actionable Comments',
            data: reports.map(r => r.prMetrics?.actionableComments || 0),
            backgroundColor: bgColors[2],
            borderColor: colors[2],
            borderWidth: 1,
          },
        ],
      },
      options: chartOptions('Quality & Responsiveness'),
    });
  }

  function chartOptions(title) {
    return {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        title: {
          display: true,
          text: title,
          font: { size: 14, weight: '600' },
          color: '#206473',
        },
        legend: {
          position: 'bottom',
          labels: {
            font: { size: 11 },
            usePointStyle: true,
            pointStyle: 'rect',
          },
        },
      },
      scales: {
        y: {
          beginAtZero: true,
          ticks: { font: { size: 11 } },
          grid: { color: 'rgba(0,0,0,0.05)' },
        },
        x: {
          ticks: { font: { size: 11 } },
          grid: { display: false },
        },
      },
    };
  }

  function destroyCharts() {
    if (bugChart) { bugChart.destroy(); bugChart = null; }
    if (prChart) { prChart.destroy(); prChart = null; }
    if (qualityChart) { qualityChart.destroy(); qualityChart = null; }
  }

  onMount(() => {
    if (reports.length) buildCharts();
  });

  onDestroy(destroyCharts);

  $: if (reports.length && bugChartCanvas) {
    buildCharts();
  }
</script>

<section class="compare-charts" aria-label="Comparison Charts">
  <div class="chart-grid">
    <div class="chart-container">
      <canvas bind:this={bugChartCanvas}></canvas>
    </div>
    <div class="chart-container">
      <canvas bind:this={prChartCanvas}></canvas>
    </div>
    <div class="chart-container">
      <canvas bind:this={qualityChartCanvas}></canvas>
    </div>
  </div>

  <div class="summary-table">
    <table>
      <thead>
        <tr>
          <th>Metric</th>
          {#each reports as r}
            <th>{shortName(r.developer)}</th>
          {/each}
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>Work Items</td>
          {#each reports as r}<td>{r.totalWorkedOn}</td>{/each}
        </tr>
        <tr>
          <td>Completed</td>
          {#each reports as r}<td>{r.resolved}</td>{/each}
        </tr>
        <tr>
          <td>Bounced Back</td>
          {#each reports as r}<td>{r.reopened}</td>{/each}
        </tr>
        <tr>
          <td>Reopen Rate</td>
          {#each reports as r}
            <td>
              <span class="rate-badge" style="background: {getRateColor(r.reopenRate || 0)}; color: white">
                {(r.reopenRate || 0).toFixed(1)}%
              </span>
            </td>
          {/each}
        </tr>
        <tr>
          <td>Avg Days to Complete</td>
          {#each reports as r}<td>{r.avgResolutionDays?.toFixed(1) || 'N/A'}</td>{/each}
        </tr>
        <tr>
          <td>Priority 1</td>
          {#each reports as r}<td>{r.priority1Count}</td>{/each}
        </tr>
        <tr>
          <td>Priority 2</td>
          {#each reports as r}<td>{r.priority2Count}</td>{/each}
        </tr>
        {#if allTypes.length > 0}
        <tr class="section-break">
          <td colspan="{reports.length + 1}" class="section-header">By Work Item Type</td>
        </tr>
        {#each allTypes as type}
        <tr>
          <td>{type}</td>
          {#each reports as r}<td>{typeCountFor(r, type)}</td>{/each}
        </tr>
        {/each}
        {/if}
        <tr class="section-break">
          <td>PRs Merged</td>
          {#each reports as r}<td>{r.prMetrics?.prsMerged ?? '-'}</td>{/each}
        </tr>
        <tr>
          <td>Total Commits</td>
          {#each reports as r}<td>{r.prMetrics?.totalCommits ?? '-'}</td>{/each}
        </tr>
        <tr>
          <td>Avg PR Cycle (days)</td>
          {#each reports as r}<td>{r.prMetrics?.avgPRCycleDays?.toFixed(1) ?? '-'}</td>{/each}
        </tr>
        <tr>
          <td>Files Changed</td>
          {#each reports as r}<td>{r.prMetrics?.filesChanged ?? '-'}</td>{/each}
        </tr>
        <tr>
          <td>Actionable Comments</td>
          {#each reports as r}<td>{r.prMetrics?.actionableComments ?? '-'}</td>{/each}
        </tr>
      </tbody>
    </table>
  </div>

  <div class="ranking-section">
    <h3 class="ranking-title">Relative Ranking</h3>
    <p class="ranking-desc">Ranked across key metrics. Lower score = better overall performance.</p>
    <div class="ranking-cards">
      {#each rankings as rank, i}
        <div class="ranking-card" class:rank-first={i === 0}>
          <span class="rank-position">#{i + 1}</span>
          <span class="rank-name">{rank.name}</span>
          <span class="rank-score">{rank.score} pts</span>
          <div class="rank-details">
            {#each rank.breakdown as item}
              <span class="rank-detail-item" title="{item.metric}: ranked #{item.rank}">
                {item.metric}: #{item.rank}
              </span>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  </div>
</section>

<style>
  .compare-charts {
    background: white;
    border-radius: 10px;
    padding: 1.5rem 2rem;
    margin-bottom: 1.5rem;
    box-shadow: 0 1px 4px rgba(46, 58, 61, 0.08), 0 4px 16px rgba(46, 58, 61, 0.06);
  }

  .chart-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .chart-container {
    height: 280px;
    position: relative;
  }

  .summary-table {
    overflow-x: auto;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }

  thead th {
    text-align: left;
    padding: 0.6rem 0.75rem;
    border-bottom: 2px solid #e2e8f0;
    font-weight: 600;
    color: #206473;
  }

  tbody td {
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid #f0f0f0;
  }

  tbody td:first-child {
    font-weight: 500;
    color: #475569;
  }

  tbody tr:hover {
    background: #f8fafb;
  }

  .section-break td {
    border-top: 2px solid #e2e8f0;
  }

  .section-header {
    font-weight: 600;
    color: #206473;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    padding-top: 0.75rem;
  }

  .rate-badge {
    display: inline-block;
    font-size: 0.75rem;
    font-weight: 600;
    padding: 0.2rem 0.5rem;
    border-radius: 10px;
  }

  .ranking-section {
    margin-top: 1.5rem;
    padding-top: 1.25rem;
    border-top: 2px solid #e2e8f0;
  }

  .ranking-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: #206473;
    margin-bottom: 0.25rem;
  }

  .ranking-desc {
    font-size: 0.78rem;
    color: #666;
    margin-bottom: 1rem;
  }

  .ranking-cards {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .ranking-card {
    flex: 1;
    min-width: 160px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 1rem;
    text-align: center;
    position: relative;
  }

  .ranking-card.rank-first {
    border-color: #206473;
    background: #f0f9fa;
  }

  .rank-position {
    display: block;
    font-size: 1.5rem;
    font-weight: 700;
    color: #206473;
  }

  .rank-first .rank-position {
    color: #0f9d58;
  }

  .rank-name {
    display: block;
    font-size: 0.85rem;
    font-weight: 600;
    color: #333;
    margin: 0.25rem 0;
  }

  .rank-score {
    display: block;
    font-size: 0.75rem;
    color: #666;
    margin-bottom: 0.5rem;
  }

  .rank-details {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
    justify-content: center;
  }

  .rank-detail-item {
    font-size: 0.65rem;
    background: #f1f5f9;
    color: #475569;
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
  }
</style>
