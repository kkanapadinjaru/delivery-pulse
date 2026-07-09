<script>
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  export let developers = [];
  export let devsLoading = false;
  export let devsError = '';

  // Persisted state passed in from parent
  export let selectedDevs = ['', '', ''];
  export let searchQueries = ['', '', ''];
  export let fromDate = '';
  export let toDate = '';

  let showDropdowns = [false, false, false];
  let activePreset = '';

  const now = new Date();
  const currentYear = now.getFullYear();

  const presets = [
    { label: 'Q1', from: `${currentYear}-01-01`, to: `${currentYear}-03-31` },
    { label: 'Q2', from: `${currentYear}-04-01`, to: `${currentYear}-06-30` },
    { label: 'Q3', from: `${currentYear}-07-01`, to: `${currentYear}-09-30` },
    { label: 'Q4', from: `${currentYear}-10-01`, to: `${currentYear}-12-31` },
    { label: 'HY1', from: `${currentYear}-01-01`, to: `${currentYear}-06-30` },
    { label: 'HY2', from: `${currentYear}-07-01`, to: `${currentYear}-12-31` },
    { label: 'YTD', from: `${currentYear}-01-01`, to: now.toISOString().split('T')[0] },
  ];

  function applyPreset(preset) {
    fromDate = preset.from;
    toDate = preset.to;
    activePreset = preset.label;
  }

  function handleDateChange() {
    activePreset = '';
  }

  // Reactive filtered lists for each developer dropdown
  $: filteredDevLists = searchQueries.map(q =>
    developers.filter(dev => dev.toLowerCase().includes((q || '').toLowerCase()))
  );

  function selectDev(idx, dev) {
    selectedDevs[idx] = dev;
    searchQueries[idx] = dev;
    showDropdowns[idx] = false;
    selectedDevs = [...selectedDevs];
    searchQueries = [...searchQueries];
  }

  function handleInput(idx) {
    // Only clear selection if the user actually changed the text
    if (searchQueries[idx] !== selectedDevs[idx]) {
      selectedDevs[idx] = '';
      selectedDevs = [...selectedDevs];
    }
    showDropdowns[idx] = true;
    searchQueries = [...searchQueries];
    showDropdowns = [...showDropdowns];
  }

  function handleFocus(idx) {
    showDropdowns[idx] = true;
    showDropdowns = [...showDropdowns];
  }

  function handleBlur(idx) {
    setTimeout(() => {
      showDropdowns[idx] = false;
      showDropdowns = [...showDropdowns];
      if (!selectedDevs[idx]) {
        const match = developers.find(d => d.toLowerCase() === searchQueries[idx].toLowerCase());
        if (match) {
          selectedDevs[idx] = match;
          searchQueries[idx] = match;
          selectedDevs = [...selectedDevs];
          searchQueries = [...searchQueries];
        }
      }
    }, 200);
  }

  $: validDevs = selectedDevs.filter(d => d !== '');
  $: canCompare = validDevs.length >= 2 && fromDate && toDate;

  function handleSubmit() {
    if (!canCompare) return;
    dispatch('compare', {
      developers: validDevs,
      from: fromDate,
      to: toDate,
    });
  }
</script>

<section class="compare-panel" aria-label="Compare Developers">
  <h2>Compare Developers (select 2-3)</h2>
  <form on:submit|preventDefault={handleSubmit}>
        <div class="devs-row">
      {#each [0, 1, 2] as idx (idx)}
        <div class="dev-input">
          <label for="dev-{idx}">Developer {idx + 1}{idx < 2 ? ' *' : ' (optional)'}</label>
          {#if devsLoading}
            <input id="dev-{idx}" type="text" disabled placeholder="Loading..." />
          {:else}
            <div class="search-dropdown">
              <input
                id="dev-{idx}"
                type="text"
                value={searchQueries[idx]}
                on:input={(e) => { searchQueries[idx] = e.target.value; handleInput(idx); }}
                on:focus={() => handleFocus(idx)}
                on:blur={() => handleBlur(idx)}
                placeholder="Search developer..."
                autocomplete="off"
              />
              {#if showDropdowns[idx] && filteredDevLists[idx].length > 0}
                <ul class="dropdown-list">
                  {#each filteredDevLists[idx] as dev}
                    <li on:mousedown={() => selectDev(idx, dev)}>{dev}</li>
                  {/each}
                </ul>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    </div>

    <div class="dates-row">
      <div class="form-group">
        <label for="cmp-from">From</label>
        <input id="cmp-from" type="date" bind:value={fromDate} on:change={handleDateChange} required />
      </div>
      <div class="form-group">
        <label for="cmp-to">To</label>
        <input id="cmp-to" type="date" bind:value={toDate} on:change={handleDateChange} required />
      </div>
      <div class="form-group form-action">
        <button type="submit" disabled={!canCompare}>Compare</button>
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
        >{preset.label}</button>
      {/each}
    </div>
  </form>
</section>

<style>
  .compare-panel {
    background: white;
    border-radius: 10px;
    padding: 1.5rem 2rem;
    margin-bottom: 2rem;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  }

  h2 {
    font-size: 1.05rem;
    margin-bottom: 1rem;
    color: #333;
  }

  .devs-row {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
    margin-bottom: 1rem;
  }

  .dev-input {
    flex: 1;
    min-width: 200px;
    display: flex;
    flex-direction: column;
  }

  .dates-row {
    display: flex;
    gap: 1rem;
    align-items: flex-end;
    flex-wrap: wrap;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    min-width: 140px;
  }

  .form-action {
    flex: 0 0 auto;
  }

  label {
    font-size: 0.8rem;
    font-weight: 600;
    margin-bottom: 0.3rem;
    color: #555;
  }

  input[type="text"], input[type="date"] {
    padding: 0.55rem 0.7rem;
    border: 1px solid #d0d5dd;
    border-radius: 6px;
    font-size: 0.85rem;
    background: white;
    width: 100%;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: #206473;
    box-shadow: 0 0 0 3px rgba(32, 100, 115, 0.1);
  }

  .search-dropdown { position: relative; }

  .dropdown-list {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    max-height: 180px;
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
    padding: 0.4rem 0.7rem;
    font-size: 0.83rem;
    cursor: pointer;
    border-bottom: 1px solid #f0f0f0;
  }

  .dropdown-list li:last-child { border-bottom: none; }
  .dropdown-list li:hover { background: #e8f4f7; color: #206473; }

  .presets-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.75rem;
    flex-wrap: wrap;
  }

  .presets-label {
    font-size: 0.78rem;
    color: #666;
    font-weight: 500;
  }

  .preset-btn {
    padding: 0.25rem 0.6rem;
    font-size: 0.75rem;
    font-weight: 500;
    background: #f1f5f9;
    color: #475569;
    border: 1px solid #e2e8f0;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .preset-btn:hover { background: #e8f4f7; border-color: #206473; color: #206473; }
  .preset-btn.active { background: #206473; color: white; border-color: #206473; }

  button[type="submit"] {
    padding: 0.55rem 1.5rem;
    background: #206473;
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }

  button[type="submit"]:hover:not(:disabled) { background: #185364; }
  button[type="submit"]:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
