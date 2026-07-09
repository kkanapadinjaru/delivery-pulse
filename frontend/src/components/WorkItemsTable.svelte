<script>
  export let workItems = [];
  export let prDetails = [];
  export let adoBaseUrl = '';

  let expanded = false;
  let sortField = 'changedDate';
  let sortDir = 'desc';

  // Build a map of work item ID -> linked PRs
  $: prByWorkItem = (() => {
    const map = {};
    for (const pr of prDetails) {
      if (pr.linkedWorkItemIds) {
        for (const wiId of pr.linkedWorkItemIds) {
          if (!map[wiId]) map[wiId] = [];
          map[wiId].push(pr);
        }
      }
    }
    return map;
  })();

  // Detect rework PRs: work items with multiple PRs to the same repo
  $: reworkByWorkItem = (() => {
    const map = {};
    for (const [wiId, prs] of Object.entries(prByWorkItem)) {
      // Group PRs by repository
      const byRepo = {};
      for (const pr of prs) {
        const repo = pr.repository || 'Unknown';
        if (!byRepo[repo]) byRepo[repo] = [];
        byRepo[repo].push(pr);
      }
      // Flag repos with more than 1 PR
      const reworkRepos = Object.entries(byRepo).filter(([, rPrs]) => rPrs.length > 1);
      if (reworkRepos.length > 0) {
        map[wiId] = reworkRepos.map(([repo, rPrs]) => ({ repo, count: rPrs.length }));
      }
    }
    return map;
  })();

  // Count of work items that have rework PRs
  $: reworkCount = Object.keys(reworkByWorkItem).length;

  // PRs that aren't linked to any work item in our list
  $: unlinkedPRs = prDetails.filter(pr => {
    if (!pr.linkedWorkItemIds || pr.linkedWorkItemIds.length === 0) return true;
    return !pr.linkedWorkItemIds.some(id => workItems.find(wi => wi.id === id));
  });

  $: sortedItems = [...workItems].sort((a, b) => {
    let aVal = a[sortField];
    let bVal = b[sortField];
    if (typeof aVal === 'string') aVal = aVal.toLowerCase();
    if (typeof bVal === 'string') bVal = bVal.toLowerCase();
    if (aVal < bVal) return sortDir === 'asc' ? -1 : 1;
    if (aVal > bVal) return sortDir === 'asc' ? 1 : -1;
    return 0;
  });

  function sort(field) {
    if (sortField === field) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc';
    } else {
      sortField = field;
      sortDir = 'asc';
    }
  }

  function stateClass(state) {
    switch (state) {
      case 'Active': case 'In Progress': return 'state-active';
      case 'Resolved': case 'Done': return 'state-resolved';
      case 'Closed': return 'state-closed';
      case 'New': case 'To Do': return 'state-new';
      case 'Removed': return 'state-removed';
      default: return '';
    }
  }

  function priorityClass(priority) {
    return priority === 1 ? 'priority-1' : 'priority-2';
  }

  function formatDate(dateStr) {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleDateString();
  }

  function workItemUrl(id) {
    if (!adoBaseUrl) return '#';
    return `${adoBaseUrl}/_workitems/edit/${id}`;
  }

  function prUrl(pr) {
    if (!adoBaseUrl) return '#';
    return `${adoBaseUrl}/_git/${pr.repository}/pullrequest/${pr.id}`;
  }

  function toggleExpand() {
    expanded = !expanded;
  }
</script>

<section class="table-section" aria-label="Details">
  <button class="section-toggle" on:click={toggleExpand} aria-expanded={expanded}>
    <span class="toggle-icon">{expanded ? '\u25BC' : '\u25B6'}</span>
    <h2>Details ({workItems.length} work items, {prDetails.length} PRs{reworkCount > 0 ? `, ${reworkCount} with rework PRs` : ''})</h2>
  </button>

  {#if expanded}
    {#if workItems.length === 0 && prDetails.length === 0}
      <p class="empty">No work items or PRs found for the selected criteria.</p>
    {:else}
      <div class="table-wrapper">
        <table>
          <thead>
            <tr>
              <th>
                <button class="sort-btn" on:click={() => sort('id')}>
                  ID {sortField === 'id' ? (sortDir === 'asc' ? '\u25B2' : '\u25BC') : ''}
                </button>
              </th>
              <th>
                <button class="sort-btn" on:click={() => sort('title')}>
                  Title {sortField === 'title' ? (sortDir === 'asc' ? '\u25B2' : '\u25BC') : ''}
                </button>
              </th>
              <th>
                <button class="sort-btn" on:click={() => sort('priority')}>
                  Pri {sortField === 'priority' ? (sortDir === 'asc' ? '\u25B2' : '\u25BC') : ''}
                </button>
              </th>
              <th>
                <button class="sort-btn" on:click={() => sort('state')}>
                  State {sortField === 'state' ? (sortDir === 'asc' ? '\u25B2' : '\u25BC') : ''}
                </button>
              </th>
              <th>Linked PR</th>
              <th>Rework</th>
              <th>
                <button class="sort-btn" on:click={() => sort('changedDate')}>
                  Last Changed {sortField === 'changedDate' ? (sortDir === 'asc' ? '\u25B2' : '\u25BC') : ''}
                </button>
              </th>
              <th>Bounced Back</th>
            </tr>
          </thead>
          <tbody>
            {#each sortedItems as item}
              <tr>
                <td class="id-col"><a href={workItemUrl(item.id)} target="_blank" rel="noopener">#{item.id}</a></td>
                <td class="title-col" title={item.title}>{item.title}</td>
                <td><span class="priority-badge {priorityClass(item.priority)}">P{item.priority}</span></td>
                <td><span class="state-badge {stateClass(item.state)}">{item.state}</span></td>
                <td class="pr-col">
                  {#if prByWorkItem[item.id]}
                    {#each prByWorkItem[item.id] as pr}
                      <a href={prUrl(pr)} target="_blank" rel="noopener" class="pr-link" title={pr.title}>PR#{pr.id}</a>
                    {/each}
                  {:else}
                    <span class="no-pr">-</span>
                  {/if}
                </td>
                <td>
                  {#if reworkByWorkItem[item.id]}
                    {#each reworkByWorkItem[item.id] as rw}
                      <span class="rework-badge" title="{rw.count} PRs to {rw.repo}">{rw.count}x {rw.repo}</span>
                    {/each}
                  {:else}
                    <span class="no-pr">-</span>
                  {/if}
                </td>
                <td>{formatDate(item.changedDate)}</td>
                <td>
                  {#if item.reactivatedCount > 0}
                    <span class="reopen-badge">{item.reactivatedCount}x</span>
                  {:else}
                    <span class="no-reopen">-</span>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      {#if unlinkedPRs.length > 0}
        <h3 class="subsection-title">PRs without linked work items ({unlinkedPRs.length})</h3>
        <div class="table-wrapper">
          <table>
            <thead>
              <tr>
                <th>PR</th>
                <th>Title</th>
                <th>Repository</th>
                <th>Status</th>
                <th>Created</th>
                <th>Closed</th>
              </tr>
            </thead>
            <tbody>
              {#each unlinkedPRs as pr}
                <tr>
                  <td class="id-col"><a href={prUrl(pr)} target="_blank" rel="noopener">PR#{pr.id}</a></td>
                  <td class="title-col" title={pr.title}>{pr.title}</td>
                  <td class="repo-col">{pr.repository}</td>
                  <td><span class="state-badge {pr.status === 'completed' ? 'state-resolved' : 'state-removed'}">{pr.status}</span></td>
                  <td>{formatDate(pr.createdDate)}</td>
                  <td>{formatDate(pr.closedDate)}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    {/if}
  {/if}
</section>

<style>
  .table-section {
    background: white;
    border-radius: 10px;
    padding: 1.5rem 2rem;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  }

  .section-toggle {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: none;
    border: none;
    cursor: pointer;
    width: 100%;
    text-align: left;
    padding: 0;
  }

  .section-toggle:hover h2 {
    color: #206473;
  }

  .toggle-icon {
    font-size: 0.8rem;
    color: #666;
  }

  h2 {
    font-size: 1.1rem;
    color: #333;
    margin: 0;
  }

  .subsection-title {
    font-size: 0.95rem;
    color: #555;
    margin-top: 1.5rem;
    margin-bottom: 0.75rem;
    padding-top: 1rem;
    border-top: 1px solid #e2e8f0;
  }

  .empty {
    color: #888;
    padding: 2rem;
    text-align: center;
  }

  .table-wrapper {
    overflow-x: auto;
    margin-top: 1rem;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }

  th {
    text-align: left;
    padding: 0.6rem 0.5rem;
    border-bottom: 2px solid #e2e8f0;
    white-space: nowrap;
  }

  .sort-btn {
    background: none;
    border: none;
    font-weight: 600;
    color: #555;
    cursor: pointer;
    padding: 0;
    font-size: 0.85rem;
  }

  .sort-btn:hover {
    color: #206473;
  }

  td {
    padding: 0.5rem 0.5rem;
    border-bottom: 1px solid #f0f0f0;
  }

  .id-col {
    white-space: nowrap;
  }

  .id-col a {
    font-weight: 600;
    color: #206473;
    text-decoration: none;
  }

  .id-col a:hover {
    text-decoration: underline;
    color: #185364;
  }

  .title-col {
    max-width: 280px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .repo-col {
    font-size: 0.8rem;
    color: #666;
    max-width: 180px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .pr-col {
    display: flex;
    gap: 0.25rem;
    flex-wrap: wrap;
  }

  .pr-link {
    font-size: 0.75rem;
    padding: 0.15rem 0.4rem;
    background: #e8f4f7;
    color: #206473;
    border-radius: 3px;
    font-weight: 500;
    white-space: nowrap;
    text-decoration: none;
  }

  .pr-link:hover {
    background: #d4edf2;
    text-decoration: underline;
  }

  .no-pr {
    color: #aaa;
  }

  .priority-badge {
    font-size: 0.75rem;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    font-weight: 600;
  }

  .priority-1 {
    background: #fde8e8;
    color: #c62828;
  }

  .priority-2 {
    background: #fff3e0;
    color: #e65100;
  }

  .state-badge {
    font-size: 0.75rem;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    font-weight: 500;
  }

  .state-active {
    background: #e3f2fd;
    color: #1565c0;
  }

  .state-resolved {
    background: #e8f5e9;
    color: #2e7d32;
  }

  .state-closed {
    background: #f3e5f5;
    color: #6a1b9a;
  }

  .state-new {
    background: #fce4ec;
    color: #ad1457;
  }

  .state-removed {
    background: #f5f5f5;
    color: #757575;
  }

  .reopen-badge {
    background: #fde8e8;
    color: #c62828;
    font-size: 0.75rem;
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
    font-weight: 600;
  }

  .rework-badge {
    display: inline-block;
    background: #fff3e0;
    color: #e65100;
    font-size: 0.7rem;
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
    font-weight: 600;
    white-space: nowrap;
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: help;
  }

  .no-reopen {
    color: #aaa;
  }

  tr:hover {
    background: #f8fafc;
  }
</style>
