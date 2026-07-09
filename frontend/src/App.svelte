<script>
  import Header from './components/Header.svelte';
  import FilterPanel from './components/FilterPanel.svelte';
  import ReportSummary from './components/ReportSummary.svelte';
  import WorkItemsTable from './components/WorkItemsTable.svelte';
  import QualityIndicators from './components/QualityIndicators.svelte';
  import HelpModal from './components/HelpModal.svelte';
  import ComparePanel from './components/ComparePanel.svelte';
  import CompareCharts from './components/CompareCharts.svelte';
  import ActivityTimeline from './components/ActivityTimeline.svelte';
  import ReposPanel from './components/ReposPanel.svelte';
  import SettingsPage from './components/SettingsPage.svelte';
  import { fetchDevelopers, fetchReport, fetchWorkItems } from './api.js';

  let page = 'main'; // 'main' or 'settings'
  let mode = 'single'; // 'single' or 'compare'
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
    } catch (e) {
      devsError = 'Could not load developers. Check backend connection.';
    } finally {
      devsLoading = false;
    }
  })();

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
</script>

<main>
  <Header on:openhelp={() => showHelp = true} on:opensettings={() => page = 'settings'} />
  <div class="container">
    {#if page === 'settings'}
      <SettingsPage on:back={() => page = 'main'} on:saved={handleSettingsSaved} />
    {:else}
    <div class="mode-toggle">
      <button class:active={mode === 'single'} on:click={() => mode = 'single'}>Single Report</button>
      <button class:active={mode === 'compare'} on:click={() => mode = 'compare'}>Compare</button>
    </div>

    {#if mode === 'single'}
      <FilterPanel
        {developers} {devsLoading} {devsError}
        bind:selectedDeveloper={filterSelectedDeveloper}
        bind:searchQuery={filterSearchQuery}
        bind:fromDate={filterFromDate}
        bind:toDate={filterToDate}
        on:search={handleSearch}
      />

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
        <QualityIndicators {report} prDetails={report.prDetails || []} />
        {#if report.prDetails && report.prDetails.length > 0}
          <ActivityTimeline prDetails={report.prDetails} from={report.from} to={report.to} />
          <ReposPanel prDetails={report.prDetails} />
        {/if}
        <WorkItemsTable {workItems} prDetails={report.prDetails || []} adoBaseUrl={report.adoBaseUrl || ''} />
      {/if}
    {:else}
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
    {/if}
    {/if}
  </div>

  {#if showHelp}
    <HelpModal on:close={() => showHelp = false} />
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
</style>
