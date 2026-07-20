<script>
  import { onMount, onDestroy } from 'svelte';
  import Chart from 'chart.js/auto';

  export let prDetails = [];
  export let throughputTrend = [];
  export let from = '';
  export let to = '';

  let chartCanvas;
  let chart;

  // Aggregate PRs merged by week (Monday-aligned) to match throughput frequency
  $: weeklyData = computeWeeklyData(prDetails, throughputTrend, from, to);

  function computeWeeklyData(prs, throughput, fromDate, toDate) {
    if (!fromDate || !toDate) return { labels: [], prsMerged: [], itemsCompleted: [] };

    // Use throughput trend weeks as the canonical week list
    const weeks = (throughput || []).map(w => w.weekStart);
    if (weeks.length === 0) return { labels: [], prsMerged: [], itemsCompleted: [] };

    // Count PRs merged per week
    const prByWeek = {};
    for (const w of weeks) prByWeek[w] = 0;

    for (const pr of prs) {
      if (pr.status === 'completed' && pr.closedDate) {
        const closedDate = new Date(pr.closedDate);
        // Find the week this PR belongs to
        for (let i = 0; i < weeks.length; i++) {
          const weekStart = new Date(weeks[i]);
          const weekEnd = new Date(weekStart);
          weekEnd.setDate(weekEnd.getDate() + 7);
          if (closedDate >= weekStart && closedDate < weekEnd) {
            prByWeek[weeks[i]]++;
            break;
          }
        }
      }
    }

    return {
      labels: weeks.map(w => formatWeekLabel(w)),
      prsMerged: weeks.map(w => prByWeek[w] || 0),
      itemsCompleted: (throughput || []).map(w => w.count),
    };
  }

  function formatWeekLabel(dateStr) {
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  function buildChart() {
    if (chart) { chart.destroy(); chart = null; }
    if (!chartCanvas || !weeklyData.labels.length) return;

    chart = new Chart(chartCanvas, {
      type: 'bar',
      data: {
        labels: weeklyData.labels,
        datasets: [
          {
            label: 'Items Completed',
            data: weeklyData.itemsCompleted,
            backgroundColor: 'rgba(32, 100, 115, 0.7)',
            borderColor: '#206473',
            borderWidth: 1,
            order: 2,
          },
          {
            label: 'PRs Merged',
            data: weeklyData.prsMerged,
            type: 'line',
            borderColor: '#d95a32',
            backgroundColor: 'rgba(217, 90, 50, 0.1)',
            fill: false,
            tension: 0.3,
            pointRadius: 3,
            pointHoverRadius: 6,
            pointBackgroundColor: '#d95a32',
            borderWidth: 2,
            order: 1,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: 'Weekly Activity — Work Items Completed vs PRs Merged',
            font: { size: 13, weight: '600' },
            color: '#206473',
          },
          legend: {
            position: 'bottom',
            labels: {
              font: { size: 11 },
              usePointStyle: true,
            },
          },
          tooltip: {
            mode: 'index',
            intersect: false,
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
            title: {
              display: true,
              text: 'Count',
              font: { size: 11 },
              color: '#666',
            },
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
    if (weeklyData.labels.length) buildChart();
  });

  onDestroy(() => {
    if (chart) { chart.destroy(); chart = null; }
  });

  $: if (weeklyData.labels.length && chartCanvas) {
    buildChart();
  }
</script>

<section class="activity-timeline" aria-label="Weekly Activity">
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

  .chart-container {
    height: 280px;
    position: relative;
  }
</style>
