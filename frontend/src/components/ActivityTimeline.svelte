<script>
  import { onMount, onDestroy } from 'svelte';
  import Chart from 'chart.js/auto';

  export let prDetails = [];
  export let from = '';
  export let to = '';

  let chartCanvas;
  let chart;

  // Parse PR dates and aggregate by day
  $: dailyData = computeDailyActivity(prDetails, from, to);

  function computeDailyActivity(prs, fromDate, toDate) {
    if (!prs.length || !fromDate || !toDate) return { labels: [], merged: [] };

    // Build a map of dates in the range
    const start = new Date(fromDate);
    const end = new Date(toDate);
    const mergedByDay = {};

    // Initialize all days in range
    const current = new Date(start);
    while (current <= end) {
      const key = current.toISOString().split('T')[0];
      mergedByDay[key] = 0;
      current.setDate(current.getDate() + 1);
    }

    // Count merged PRs per day (by closedDate for completed PRs)
    for (const pr of prs) {
      if (pr.status === 'completed' && pr.closedDate) {
        const day = pr.closedDate.split('T')[0];
        if (mergedByDay[day] !== undefined) {
          mergedByDay[day]++;
        }
      }
    }

    const sortedDays = Object.keys(mergedByDay).sort();
    return {
      labels: sortedDays.map(d => formatDateLabel(d)),
      merged: sortedDays.map(d => mergedByDay[d]),
    };
  }

  function formatDateLabel(dateStr) {
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  function buildChart() {
    if (chart) { chart.destroy(); chart = null; }
    if (!chartCanvas || !dailyData.labels.length) return;

    chart = new Chart(chartCanvas, {
      type: 'line',
      data: {
        labels: dailyData.labels,
        datasets: [
          {
            label: 'PRs Merged',
            data: dailyData.merged,
            borderColor: '#206473',
            backgroundColor: 'rgba(32, 100, 115, 0.1)',
            fill: true,
            tension: 0.3,
            pointRadius: 2,
            pointHoverRadius: 5,
            pointBackgroundColor: '#206473',
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: 'Daily PR Activity',
            font: { size: 14, weight: '600' },
            color: '#206473',
          },
          legend: {
            position: 'bottom',
            labels: {
              font: { size: 11 },
              usePointStyle: true,
            },
          },
        },
        scales: {
          y: {
            beginAtZero: true,
            ticks: {
              stepSize: 1,
              font: { size: 11 },
            },
            grid: { color: 'rgba(0,0,0,0.05)' },
          },
          x: {
            ticks: {
              font: { size: 10 },
              maxRotation: 45,
              autoSkip: true,
              maxTicksLimit: 20,
            },
            grid: { display: false },
          },
        },
      },
    });
  }

  onMount(() => {
    if (dailyData.labels.length) buildChart();
  });

  onDestroy(() => {
    if (chart) { chart.destroy(); chart = null; }
  });

  $: if (dailyData.labels.length && chartCanvas) {
    buildChart();
  }
</script>

<section class="activity-timeline" aria-label="Activity Timeline">
  <div class="panel-header">
    <h3>Activity Timeline</h3>
  </div>
  <div class="chart-container">
    <canvas bind:this={chartCanvas}></canvas>
  </div>
</section>

<style>
  .activity-timeline {
    background: white;
    border-radius: 10px;
    padding: 1.5rem 2rem;
    margin-bottom: 1.5rem;
    box-shadow: 0 1px 4px rgba(46, 58, 61, 0.08), 0 4px 16px rgba(46, 58, 61, 0.06);
  }

  .panel-header h3 {
    font-size: 0.95rem;
    color: #206473;
    margin: 0 0 1rem 0;
    font-weight: 600;
  }

  .chart-container {
    height: 260px;
    position: relative;
  }
</style>
