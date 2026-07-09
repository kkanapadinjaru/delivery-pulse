<script>
  import { createEventDispatcher } from 'svelte';
  import { fetchDevelopers } from '../api.js';

  const dispatch = createEventDispatcher();

  export let developers = [];
  export let devsLoading = false;
  export let devsError = '';

  // Persisted state passed in from parent
  export let selectedDeveloper = '';
  export let searchQuery = '';
  export let fromDate = '';
  export let toDate = '';

  let showDropdown = false;
  let activePreset = '';

  // Date presets based on current year
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

  function applyPreset(preset) {
    fromDate = preset.from;
    toDate = preset.to;
    activePreset = preset.label;
  }

  function handleDateChange() {
    activePreset = '';
  }

  $: filteredDevelopers = developers.filter(dev =>
    dev.toLowerCase().includes(searchQuery.toLowerCase())
  );

  function selectDeveloper(dev) {
    selectedDeveloper = dev;
    searchQuery = dev;
    showDropdown = false;
  }

  function handleInput() {
    selectedDeveloper = '';
    showDropdown = true;
  }

  function handleFocus() {
    showDropdown = true;
  }

  function handleBlur() {
    setTimeout(() => {
      showDropdown = false;
      if (!selectedDeveloper) {
        const match = developers.find(d => d.toLowerCase() === searchQuery.toLowerCase());
        if (match) {
          selectedDeveloper = match;
          searchQuery = match;
        }
      }
    }, 200);
  }

  function handleSubmit() {
    if (!selectedDeveloper || !fromDate || !toDate) return;
    dispatch('search', {
      developer: selectedDeveloper,
      from: fromDate,
      to: toDate,
    });
  }
</script>

<section class="filter-panel" aria-label="Report Filters">
  <h2>Generate Report</h2>
  <form on:submit|preventDefault={handleSubmit}>
    <div class="form-row">
      <div class="form-group developer-group">
        <label for="developer">Developer</label>
        {#if devsLoading}
          <input id="developer" type="text" disabled placeholder="Loading developers..." />
        {:else if devsError}
          <input
            id="developer"
            type="text"
            bind:value={searchQuery}
            on:input={handleInput}
            placeholder="Enter developer name or email"
          />
          <small class="field-error">{devsError}</small>
        {:else}
          <div class="search-dropdown">
            <input
              id="developer"
              type="text"
              bind:value={searchQuery}
              on:input={handleInput}
              on:focus={handleFocus}
              on:blur={handleBlur}
              placeholder="Search developer..."
              autocomplete="off"
              role="combobox"
              aria-expanded={showDropdown && filteredDevelopers.length > 0}
              aria-controls="developer-listbox"
            />
            {#if showDropdown && filteredDevelopers.length > 0}
              <ul class="dropdown-list" id="developer-listbox" role="listbox">
                {#each filteredDevelopers as dev}
                  <li
                    role="option"
                    aria-selected={dev === selectedDeveloper}
                    on:mousedown={() => selectDeveloper(dev)}
                  >
                    {dev}
                  </li>
                {/each}
              </ul>
            {/if}
          </div>
        {/if}
      </div>

      <div class="form-group">
        <label for="from-date">From</label>
        <input id="from-date" type="date" bind:value={fromDate} on:change={handleDateChange} required />
      </div>

      <div class="form-group">
        <label for="to-date">To</label>
        <input id="to-date" type="date" bind:value={toDate} on:change={handleDateChange} required />
      </div>

      <div class="form-group form-action">
        <button type="submit" disabled={!selectedDeveloper || !fromDate || !toDate}>
          Generate Report
        </button>
      </div>
    </div>

    <div class="presets-row">
      <span class="presets-label">Quick select:</span>
      {#each presets as preset}
        <button
          type="button"
          class="preset-btn"
          class:active={activePreset === preset.label}
          on:click={() => applyPreset(preset)}
        >
          {preset.label}
        </button>
      {/each}
    </div>
  </form>
</section>

<style>
  .filter-panel {
    background: white;
    border-radius: 10px;
    padding: 1.5rem 2rem;
    margin-bottom: 2rem;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  }

  h2 {
    font-size: 1.1rem;
    margin-bottom: 1rem;
    color: #333;
  }

  .form-row {
    display: flex;
    gap: 1rem;
    align-items: flex-end;
    flex-wrap: wrap;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 180px;
  }

  .developer-group {
    flex: 2;
    min-width: 280px;
  }

  .form-action {
    flex: 0 0 auto;
  }

  label {
    font-size: 0.85rem;
    font-weight: 600;
    margin-bottom: 0.35rem;
    color: #555;
  }

  input[type="text"], input[type="date"] {
    padding: 0.6rem 0.75rem;
    border: 1px solid #d0d5dd;
    border-radius: 6px;
    font-size: 0.9rem;
    background: white;
    width: 100%;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: #206473;
    box-shadow: 0 0 0 3px rgba(32, 100, 115, 0.1);
  }

  .search-dropdown {
    position: relative;
  }

  .dropdown-list {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    max-height: 220px;
    overflow-y: auto;
    background: white;
    border: 1px solid #d0d5dd;
    border-top: none;
    border-radius: 0 0 6px 6px;
    list-style: none;
    margin: 0;
    padding: 0;
    z-index: 100;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .dropdown-list li {
    padding: 0.5rem 0.75rem;
    font-size: 0.9rem;
    cursor: pointer;
    border-bottom: 1px solid #f0f0f0;
  }

  .dropdown-list li:last-child {
    border-bottom: none;
  }

  .dropdown-list li:hover {
    background: #e8f4f7;
    color: #206473;
  }

  .dropdown-list li[aria-selected="true"] {
    background: #d4edf2;
    font-weight: 600;
  }

  .presets-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 1rem;
    flex-wrap: wrap;
  }

  .presets-label {
    font-size: 0.8rem;
    color: #666;
    font-weight: 500;
  }

  .preset-btn {
    padding: 0.3rem 0.7rem;
    font-size: 0.78rem;
    font-weight: 500;
    background: #f1f5f9;
    color: #475569;
    border: 1px solid #e2e8f0;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .preset-btn:hover {
    background: #e8f4f7;
    border-color: #206473;
    color: #206473;
  }

  .preset-btn.active {
    background: #206473;
    color: white;
    border-color: #206473;
  }

  button[type="submit"] {
    padding: 0.6rem 1.5rem;
    background: #206473;
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }

  button[type="submit"]:hover:not(:disabled) {
    background: #185364;
  }

  button[type="submit"]:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .field-error {
    color: #d32f2f;
    font-size: 0.8rem;
    margin-top: 0.25rem;
  }
</style>
