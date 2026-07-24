<script>
  import Header from './components/Header.svelte';
  import FilterPanel from './components/FilterPanel.svelte';
  import ReportSummary from './components/ReportSummary.svelte';
  import MetricsInsights from './components/MetricsInsights.svelte';
  import WorkItemsTable from './components/WorkItemsTable.svelte';
  import QualityIndicators from './components/QualityIndicators.svelte';
  import HelpModal from './components/HelpModal.svelte';
  import ComparePanel from './components/ComparePanel.svelte';
  import CompareCharts from './components/CompareCharts.svelte';
  import ActivityTimeline from './components/ActivityTimeline.svelte';
  import ReposPanel from './components/ReposPanel.svelte';
  import SettingsPage from './components/SettingsPage.svelte';
  import TeamDashboard from './components/TeamDashboard.svelte';
  import { fetchDevelopers, fetchReport, fetchWorkItems } from './api.js';
  import { logout } from './lib/keycloak.js';

  // currentUser is passed as a prop from main.js after authentication
  export let currentUser = null;

  $: isManager = currentUser?.roles?.includes('PulseManager') || false;
  $: userEmail = currentUser?.email || '';

  let page = 'main'; // 'main' or 'settings'
  let mode = 'single'; // 'single', 'compare', or 'team'
  let developers = [];
  let devsLoading = true;
  let devsError = '';

  let developer = '';
  let fromDate = '';
  let toDate = '';
  let report = null;
  let workItems = [];
  let loading = false;
  let error = '';
  let showHelp = false;

  // Filter panel persisted state
  let filterSelectedDeveloper = '';
  let filterSearchQuery = '';
  let filterFromDate = '';
  let filterToDate = '';

  // Compare panel persisted state
  let compareSelectedDevs = ['', '', ''];
  let compareSearchQueries = ['', '', ''];
  let compareFromDate = '';
  let compareToDate = '';

  // Compare mode state
  let compareReports = [];
  let compareLoading = false;
  let compareError = '';

  // Set default date ranges
  const now = new Date();
  const thirtyDaysAgo = new Date(now);
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
  filterFromDate = thirtyDaysAgo.toISOString().split('T')[0];
  filterToDate = now.toISOString().split('T')[0];
  compareFromDate = thirtyDaysAgo.toISOString().split('T')[0];
  compareToDate = now.toISOString().split('T')[0];

  // Load developers once on app init
  (async () => {
    try {
      const devs = await fetchDevelopers();
      developers = devs.sort((a, b) => a.localeCompare(b, undefined, { sensitivity: 'base' }));

      // For developer role: auto-select their own email and load report
      if (!isManager && userEmail) {
        filterSelectedDeveloper = userEmail;
        filterSearchQuery = userEmail;
        // Auto-load their report
        autoLoadDeveloperReport();
      }
    } catch (e) {
      devsError = 'Could not load developers. Check backend connection.';
    } finally {
      devsLoading = false;
    }
  })();

  async function autoLoadDeveloperReport() {
    developer = userEmail;
    fromDate = filterFromDate;
    toDate = filterToDate;
    loading = true;
    error = '';
    report = null;
    workItems = [];

    try {
      const [reportData, itemsData] = await Promise.all([
        fetchReport(userEmail, fromDate, toDate),
        fetchWorkItems(userEmail, fromDate, toDate),
      ]);
      report = reportData;
      workItems = itemsData.workItems || [];
    } catch (e) {
      error = e.message || 'Failed to fetch report';
    } finally {
      loading = false;
    }
  }

  async function handleSearch(event) {
    developer = event.detail.developer;
    fromDate = event.detail.from;
    toDate = event.detail.to;
    loading = true;
    error = '';
    report = null;
    workItems = [];

    try {
      const [reportData, itemsData] = await Promise.all([
        fetchReport(developer, fromDate, toDate),
        fetchWorkItems(developer, fromDate, toDate),
      ]);
      report = reportData;
      workItems = itemsData.workItems || [];
    } catch (e) {
      error = e.message || 'Failed to fetch report';
    } finally {
      loading = false;
    }
  }

  async function handleCompare(event) {
    const { developers: devs, from, to } = event.detail;
    compareLoading = true;
    compareError = '';
    compareReports = [];

    try {
      const promises = devs.map(dev => fetchReport(dev, from, to));
      compareReports = await Promise.all(promises);
    } catch (e) {
      compareError = e.message || 'Failed to fetch comparison data';
    } finally {
      compareLoading = false;
    }
  }

  async function handleSettingsSaved() {
    // Reload developers list since teams may have changed
    devsLoading = true;
    devsError = '';
    try {
      const devs = await fetchDevelopers();
      developers = devs.sort((a, b) => a.localeCompare(b, undefined, { sensitivity: 'base' }));
    } catch (e) {
      devsError = 'Could not load developers. Check backend connection.';
    } finally {
      devsLoading = false;
    }
  }

  function exportPDF() {
    window.print();
  }
</script>

<main>
  <Header on:openhelp={() => showHelp = true} on:opensettings={() => page = 'settings'} {currentUser} {isManager} on:logout={logout} />
  <div class="container">
    {#if page === 'settings' && isManager}
      <SettingsPage on:back={() => page = 'main'} on:saved={handleSettingsSaved} />
    {:else}
    {#if isManager}
    <div class="mode-toggle">
      <button class:active={mode === 'single'} on:click={() => mode = 'single'}>Single Report</button>
      <button class:active={mode === 'compare'} on:click={() => mode = 'compare'}>Compare</button>
      <button class:active={mode === 'team'} on:click={() => mode = 'team'}>Team</button>
    </div>
    {/if}

    {#if (mode === 'single' && report) || (mode === 'compare' && compareReports.length >= 2) || (mode === 'team')}
      <button class="export-btn" on:click={exportPDF}>Export PDF</button>
    {/if}

    {#if mode === 'single'}
      {#if isManager}
      <FilterPanel
        {developers} {devsLoading} {devsError}
        bind:selectedDeveloper={filterSelectedDeveloper}
        bind:searchQuery={filterSearchQuery}
        bind:fromDate={filterFromDate}
        bind:toDate={filterToDate}
        on:search={handleSearch}
      />
      {:else}
      <!-- Developer role: date-only filter, auto-uses their email -->
      <div class="dev-date-filter">
        <p class="dev-greeting">Your Performance Report</p>
        <div class="dev-date-controls">
          <input type="date" bind:value={filterFromDate} />
          <span class="date-sep">to</span>
          <input type="date" bind:value={filterToDate} />
          <button class="refresh-btn" on:click={autoLoadDeveloperReport}>Load Report</button>
        </div>
      </div>
      {/if}

      {#if loading}
        <div class="loading">
          <div class="spinner"></div>
          <p>Loading report data from Azure DevOps...</p>
        </div>
      {/if}

      {#if error}
        <div class="error" role="alert">
          <strong>Error:</strong> {error}
        </div>
      {/if}

      {#if report}
        <ReportSummary {report} />
        <MetricsInsights {report} />
        <QualityIndicators {report} prDetails={report.prDetails || []} />
        {#if (report.prDetails && report.prDetails.length > 0) || (report.throughputTrend && report.throughputTrend.length > 0)}
          <ActivityTimeline prDetails={report.prDetails || []} throughputTrend={report.throughputTrend || []} from={report.from} to={report.to} />
        {/if}
        {#if report.prDetails && report.prDetails.length > 0}
          <ReposPanel prDetails={report.prDetails} />
        {/if}
        <WorkItemsTable {workItems} prDetails={report.prDetails || []} adoBaseUrl={report.adoBaseUrl || ''} />
      {/if}
    {:else if mode === 'compare'}
      <ComparePanel
        {developers} {devsLoading} {devsError}
        bind:selectedDevs={compareSelectedDevs}
        bind:searchQueries={compareSearchQueries}
        bind:fromDate={compareFromDate}
        bind:toDate={compareToDate}
        on:compare={handleCompare}
      />

      {#if compareLoading}
        <div class="loading">
          <div class="spinner"></div>
          <p>Fetching reports for comparison...</p>
        </div>
      {/if}

      {#if compareError}
        <div class="error" role="alert">
          <strong>Error:</strong> {compareError}
        </div>
      {/if}

      {#if compareReports.length >= 2}
        <CompareCharts reports={compareReports} />
      {/if}
    {:else if mode === 'team'}
      <TeamDashboard />
    {/if}
    {/if}
  </div>

  {#if showHelp}
    <HelpModal on:close={() => showHelp = false} {isManager} />
  {/if}
</main>

<style>
  main {
    min-height: 100vh;
  }

  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
  }

  .mode-toggle {
    display: flex;
    gap: 0;
    margin-bottom: 1.5rem;
    background: #e2e8f0;
    border-radius: 8px;
    padding: 3px;
    width: fit-content;
  }

  .mode-toggle button {
    padding: 0.45rem 1.2rem;
    border: none;
    border-radius: 6px;
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
    background: transparent;
    color: #555;
    transition: all 0.2s;
  }

  .mode-toggle button.active {
    background: white;
    color: #206473;
    font-weight: 600;
    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  }

  .mode-toggle button:hover:not(.active) {
    color: #206473;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 3rem;
    color: #555;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid #e0e0e0;
    border-top-color: #206473;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin-bottom: 1rem;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error {
    background: #fde8e8;
    border: 1px solid #f5aca6;
    border-radius: 8px;
    padding: 1rem 1.5rem;
    margin-bottom: 1.5rem;
    color: #9b1c1c;
  }

  .export-btn {
    padding: 0.4rem 1rem;
    background: transparent;
    color: #206473;
    border: 1px solid #206473;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 500;
    cursor: pointer;
    margin-bottom: 1.5rem;
    transition: all 0.15s;
  }

  .export-btn:hover {
    background: #e8f4f7;
  }

  .dev-date-filter {
    background: white;
    border-radius: 10px;
    padding: 1.5rem 2rem;
    margin-bottom: 1.5rem;
    box-shadow: 0 1px 4px rgba(46, 58, 61, 0.08), 0 4px 16px rgba(46, 58, 61, 0.06);
  }

  .dev-greeting {
    font-family: 'Manrope', 'Inter', system-ui, sans-serif;
    font-size: 1.1rem;
    font-weight: 600;
    color: #206473;
    margin-bottom: 1rem;
  }

  .dev-date-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .dev-date-controls input {
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

  .refresh-btn:hover { background: #185364; }

  /* Print styles */
  @media print {
    :global(body) {
      background: white !important;
    }

    :global(.mode-toggle),
    :global(.export-btn),
    :global(header),
    :global(.filter-panel),
    :global(.compare-panel) {
      display: none !important;
    }

    /* Hide detail sections in single report */
    :global(.activity-timeline),
    :global(.repos-panel),
    :global(.work-items-table) {
      display: none !important;
    }

    /* Hide team dashboard controls in print */
    :global(.date-controls),
    :global(.refresh-btn) {
      display: none !important;
    }

    :global(.container) {
      max-width: 100% !important;
      padding: 0 !important;
    }

    :global(.summary),
    :global(.metrics-insights),
    :global(.quality),
    :global(.compare-charts),
    :global(.team-dashboard) {
      box-shadow: none !important;
      break-inside: avoid;
    }
  }
</style>
