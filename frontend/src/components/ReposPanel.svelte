<script>
  export let prDetails = [];

  let expanded = false;

  // Aggregate PRs by repository
  $: repoData = computeRepoData(prDetails);

  function computeRepoData(prs) {
    if (!prs.length) return [];

    const repoMap = {};
    for (const pr of prs) {
      const repo = pr.repository || 'Unknown';
      if (!repoMap[repo]) {
        repoMap[repo] = { repository: repo, prsCreated: 0, merged: 0, avgCycleDays: 0, cycleDaysTotal: 0, cycleCount: 0 };
      }
      repoMap[repo].prsCreated++;
      if (pr.status === 'completed') {
        repoMap[repo].merged++;
        if (pr.createdDate && pr.closedDate) {
          const created = new Date(pr.createdDate);
          const closed = new Date(pr.closedDate);
          const days = (closed - created) / (1000 * 60 * 60 * 24);
          repoMap[repo].cycleDaysTotal += days;
          repoMap[repo].cycleCount++;
        }
      }
    }

    return Object.values(repoMap)
      .map(r => ({
        ...r,
        avgCycleDays: r.cycleCount > 0 ? (r.cycleDaysTotal / r.cycleCount).toFixed(1) : '-',
      }))
      .sort((a, b) => b.prsCreated - a.prsCreated);
  }
</script>

<section class="repos-panel" aria-label="Repos Contributed To">
  <button class="panel-toggle" on:click={() => expanded = !expanded} aria-expanded={expanded}>
    <span class="toggle-icon">{expanded ? '▾' : '▸'}</span>
    <h3>Repos Contributed To</h3>
    <span class="badge">{repoData.length}</span>
  </button>

  {#if expanded}
    <div class="panel-content">
      {#if repoData.length === 0}
        <p class="empty">No repository data available.</p>
      {:else}
        <table>
          <thead>
            <tr>
              <th>Repository</th>
              <th>PRs Created</th>
              <th>Merged</th>
              <th>Avg Cycle Days</th>
            </tr>
          </thead>
          <tbody>
            {#each repoData as repo}
              <tr>
                <td class="repo-name">{repo.repository}</td>
                <td>{repo.prsCreated}</td>
                <td>{repo.merged}</td>
                <td>{repo.avgCycleDays}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </div>
  {/if}
</section>

<style>
  .repos-panel {
    background: white;
    border-radius: 10px;
    padding: 0;
    margin-bottom: 1.5rem;
    box-shadow: 0 1px 4px rgba(46, 58, 61, 0.08), 0 4px 16px rgba(46, 58, 61, 0.06);
    overflow: hidden;
  }

  .panel-toggle {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    padding: 1rem 1.5rem;
    background: none;
    border: none;
    cursor: pointer;
    text-align: left;
    transition: background 0.15s;
  }

  .panel-toggle:hover {
    background: #f8fafb;
  }

  .toggle-icon {
    font-size: 0.85rem;
    color: #666;
    width: 1rem;
  }

  .panel-toggle h3 {
    font-size: 0.95rem;
    color: #206473;
    margin: 0;
    font-weight: 600;
    flex: 1;
  }

  .badge {
    background: #e8f4f7;
    color: #206473;
    font-size: 0.75rem;
    font-weight: 600;
    padding: 0.15rem 0.5rem;
    border-radius: 10px;
  }

  .panel-content {
    padding: 0 1.5rem 1.5rem;
  }

  .empty {
    color: #666;
    font-size: 0.85rem;
    font-style: italic;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }

  thead th {
    text-align: left;
    padding: 0.5rem 0.75rem;
    border-bottom: 2px solid #e2e8f0;
    font-weight: 600;
    color: #206473;
    font-size: 0.8rem;
  }

  tbody td {
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid #f0f0f0;
  }

  tbody tr:hover {
    background: #f8fafb;
  }

  .repo-name {
    font-weight: 500;
    color: #333;
  }
</style>
