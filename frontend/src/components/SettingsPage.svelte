<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { getAuthHeaders } from '../api.js';

  const dispatch = createEventDispatcher();

  // Standard Azure DevOps work item types
  const allWorkItemTypes = [
    'Bug',
    'Task',
    'User Story',
    'Feature',
    'Epic',
    'Issue',
    'Test Case',
    'Test Plan',
    'Test Suite',
    'Impediment',
    'Change Request',
    'Risk',
    'Review',
  ];

  // Common ADO activity types
  const allActivityTypes = [
    'Development',
    'Testing',
    'Requirements',
    'Design',
    'Deployment',
    'Documentation',
  ];

  let teams = '';
  let selectedTypes = [];
  let selectedAreaPaths = [];
  let selectedActivities = [];
  let selectedDevelopers = [];
  let availableAreaPaths = [];
  let availableDevelopers = [];
  let devSearchQuery = '';
  let newActivityInput = '';
  let prSizeSmallMax = 25;
  let prSizeMediumMax = 100;
  let loading = true;
  let saving = false;
  let error = '';
  let success = '';
  let showTypeDropdown = false;
  let showAreaPathDropdown = false;
  let showActivityDropdown = false;
  let showDevDropdown = false;
  let devSearchInput;
  let loadingAreaPaths = false;

  onMount(async () => {
    try {
      const [settingsResp, areaPathsResp, allDevsResp] = await Promise.all([
        fetch('/api/settings', { headers: getAuthHeaders() }),
        fetch('/api/areapaths', { headers: getAuthHeaders() }),
        fetch('/api/all-developers', { headers: getAuthHeaders() }),
      ]);

      if (!settingsResp.ok) throw new Error('Failed to load settings');
      const data = await settingsResp.json();
      teams = (data.teams || []).join(', ');
      selectedTypes = data.workItemTypes || ['Bug', 'Task'];
      selectedAreaPaths = data.areaPaths || [];
      selectedActivities = data.activities || ['Development', 'Testing', 'Requirements'];
      selectedDevelopers = data.developers || [];
      prSizeSmallMax = data.prSizeSmallMax || 25;
      prSizeMediumMax = data.prSizeMediumMax || 100;

      if (areaPathsResp.ok) {
        const apData = await areaPathsResp.json();
        availableAreaPaths = apData.areaPaths || [];
      }

      if (allDevsResp.ok) {
        const devData = await allDevsResp.json();
        availableDevelopers = (devData.developers || []).sort((a, b) => a.localeCompare(b, undefined, { sensitivity: 'base' }));
      }
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });

  function toggleType(type) {
    if (selectedTypes.includes(type)) {
      selectedTypes = selectedTypes.filter(t => t !== type);
    } else {
      selectedTypes = [...selectedTypes, type];
    }
  }

  function removeType(type) {
    selectedTypes = selectedTypes.filter(t => t !== type);
  }

  function toggleAreaPath(path) {
    if (selectedAreaPaths.includes(path)) {
      selectedAreaPaths = selectedAreaPaths.filter(p => p !== path);
    } else {
      selectedAreaPaths = [...selectedAreaPaths, path];
    }
  }

  function removeAreaPath(path) {
    selectedAreaPaths = selectedAreaPaths.filter(p => p !== path);
  }

  function toggleActivity(activity) {
    if (selectedActivities.includes(activity)) {
      selectedActivities = selectedActivities.filter(a => a !== activity);
    } else {
      selectedActivities = [...selectedActivities, activity];
    }
  }

  function removeActivity(activity) {
    selectedActivities = selectedActivities.filter(a => a !== activity);
  }

  function addCustomActivity() {
    const val = newActivityInput.trim();
    if (val && !selectedActivities.includes(val)) {
      selectedActivities = [...selectedActivities, val];
    }
    newActivityInput = '';
  }

  function toggleDeveloper(dev) {
    if (selectedDevelopers.includes(dev)) {
      selectedDevelopers = selectedDevelopers.filter(d => d !== dev);
    } else {
      selectedDevelopers = [...selectedDevelopers, dev];
    }
  }

  function removeDeveloper(dev) {
    selectedDevelopers = selectedDevelopers.filter(d => d !== dev);
  }

  function shortEmail(email) {
    const local = email.split('@')[0] || email;
    return local.split('.').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ');
  }

  $: filteredAvailableDevs = availableDevelopers
    .filter(d => !selectedDevelopers.includes(d))
    .filter(d => {
      if (!devSearchQuery) return true;
      const q = devSearchQuery.toLowerCase();
      return d.toLowerCase().includes(q) || shortEmail(d).toLowerCase().includes(q);
    });

  function handleActivityKeydown(e) {
    if (e.key === 'Enter') {
      e.preventDefault();
      addCustomActivity();
    }
  }

  async function handleSave() {
    saving = true;
    error = '';
    success = '';

    const teamsList = teams
      .split(',')
      .map(t => t.trim())
      .filter(t => t !== '');

    try {
      const resp = await fetch('/api/settings', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify({
          teams: teamsList,
          developers: selectedDevelopers,
          workItemTypes: selectedTypes,
          areaPaths: selectedAreaPaths,
          activities: selectedActivities,
          prSizeSmallMax: parseInt(prSizeSmallMax) || 25,
          prSizeMediumMax: parseInt(prSizeMediumMax) || 100,
        }),
      });

      if (!resp.ok) {
        const data = await resp.json().catch(() => ({}));
        throw new Error(data.error || 'Failed to save settings');
      }

      success = 'Settings saved successfully. Changes will take effect on the next report generation.';
      dispatch('saved');
    } catch (e) {
      error = e.message;
    } finally {
      saving = false;
    }
  }

  function handleDropdownBlur(setter) {
    setTimeout(() => { setter(false); }, 150);
  }

  // Get the short display name for an area path (last segment)
  function shortAreaPath(path) {
    const parts = path.split('\\');
    return parts.length > 1 ? parts.slice(1).join(' > ') : path;
  }
</script>

<section class="settings-page" aria-label="Application Settings">
  <h2>Settings</h2>

  {#if loading}
    <div class="loading-state">Loading settings...</div>
  {:else}
    <form on:submit|preventDefault={handleSave}>
      <div class="setting-group">
        <label for="teams">ADO Teams</label>
        <p class="help-text">Comma-separated list of Azure DevOps team names to load developers from. Leave empty to use all teams in the project.</p>
        <input
          id="teams"
          type="text"
          bind:value={teams}
          placeholder="e.g. Team Alpha, Team Beta"
        />
      </div>

      <div class="setting-group">
        <label>Developers</label>
        <p class="help-text">Select which team members to include in reports. Leave empty to include all members from the configured teams.</p>

        <div class="selected-types">
          {#each selectedDevelopers as dev}
            <span class="type-tag" title={dev}>
              {shortEmail(dev)}
              <button type="button" class="remove-tag" on:click={() => removeDeveloper(dev)} aria-label="Remove {dev}">&times;</button>
            </span>
          {/each}
          {#if selectedDevelopers.length === 0}
            <span class="no-types">No filter (all team members included)</span>
          {/if}
        </div>

        <div class="type-dropdown-wrapper">
          <button
            type="button"
            class="dropdown-trigger"
            on:click={() => { showDevDropdown = !showDevDropdown; devSearchQuery = ''; setTimeout(() => devSearchInput?.focus(), 0); }}
          >
            Add developer...
          </button>
          {#if showDevDropdown}
            <div class="type-dropdown area-dropdown" on:mouseleave={() => {}} on:focusout={(e) => { if (!e.currentTarget.contains(e.relatedTarget)) showDevDropdown = false; }}>
              <div class="search-input-item">
                <input
                  type="text"
                  bind:this={devSearchInput}
                  bind:value={devSearchQuery}
                  placeholder="Search by name or email..."
                  class="dropdown-search"
                />
              </div>
              <ul class="dropdown-list-inner">
                {#each filteredAvailableDevs as dev}
                  <li on:mousedown={() => toggleDeveloper(dev)} title={dev}>{shortEmail(dev)} <span class="dev-email">({dev})</span></li>
                {/each}
                {#if filteredAvailableDevs.length === 0}
                  <li class="empty-item">{devSearchQuery ? 'No matches' : 'All developers selected'}</li>
                {/if}
              </ul>
            </div>
          {/if}
        </div>
      </div>

      <div class="setting-group">
        <label>Work Item Types</label>
        <p class="help-text">Select which work item types to include in reports. At least one type must be selected.</p>

        <div class="selected-types">
          {#each selectedTypes as type}
            <span class="type-tag">
              {type}
              <button type="button" class="remove-tag" on:click={() => removeType(type)} aria-label="Remove {type}">&times;</button>
            </span>
          {/each}
          {#if selectedTypes.length === 0}
            <span class="no-types">No types selected</span>
          {/if}
        </div>

        <div class="type-dropdown-wrapper">
          <button
            type="button"
            class="dropdown-trigger"
            on:click={() => showTypeDropdown = !showTypeDropdown}
            on:blur={() => handleDropdownBlur(v => showTypeDropdown = v)}
          >
            Add work item type...
          </button>
          {#if showTypeDropdown}
            <ul class="type-dropdown">
              {#each allWorkItemTypes.filter(t => !selectedTypes.includes(t)) as type}
                <li on:mousedown={() => toggleType(type)}>{type}</li>
              {/each}
              {#if allWorkItemTypes.filter(t => !selectedTypes.includes(t)).length === 0}
                <li class="empty-item">All types selected</li>
              {/if}
            </ul>
          {/if}
        </div>
      </div>

      <div class="setting-group">
        <label>Area Paths</label>
        <p class="help-text">Select which area paths to scope work item queries to. Leave empty to include all areas.</p>

        <div class="selected-types">
          {#each selectedAreaPaths as path}
            <span class="type-tag area-tag" title={path}>
              {shortAreaPath(path)}
              <button type="button" class="remove-tag" on:click={() => removeAreaPath(path)} aria-label="Remove {path}">&times;</button>
            </span>
          {/each}
          {#if selectedAreaPaths.length === 0}
            <span class="no-types">No area paths selected (all areas included)</span>
          {/if}
        </div>

        <div class="type-dropdown-wrapper">
          <button
            type="button"
            class="dropdown-trigger"
            on:click={() => showAreaPathDropdown = !showAreaPathDropdown}
            on:blur={() => handleDropdownBlur(v => showAreaPathDropdown = v)}
          >
            Add area path...
          </button>
          {#if showAreaPathDropdown}
            <ul class="type-dropdown area-dropdown">
              {#each availableAreaPaths.filter(p => !selectedAreaPaths.includes(p)) as path}
                <li on:mousedown={() => toggleAreaPath(path)} title={path}>{shortAreaPath(path)}</li>
              {/each}
              {#if availableAreaPaths.filter(p => !selectedAreaPaths.includes(p)).length === 0}
                <li class="empty-item">All available paths selected</li>
              {/if}
            </ul>
          {/if}
        </div>
      </div>

      <div class="setting-group">
        <label>Activity Types</label>
        <p class="help-text">Filter work items by activity type. Leave empty to include all activities.</p>

        <div class="selected-types">
          {#each selectedActivities as activity}
            <span class="type-tag">
              {activity}
              <button type="button" class="remove-tag" on:click={() => removeActivity(activity)} aria-label="Remove {activity}">&times;</button>
            </span>
          {/each}
          {#if selectedActivities.length === 0}
            <span class="no-types">No activity filter (all activities included)</span>
          {/if}
        </div>

        <div class="type-dropdown-wrapper">
          <button
            type="button"
            class="dropdown-trigger"
            on:click={() => showActivityDropdown = !showActivityDropdown}
            on:blur={() => handleDropdownBlur(v => showActivityDropdown = v)}
          >
            Add activity type...
          </button>
          {#if showActivityDropdown}
            <ul class="type-dropdown">
              {#each allActivityTypes.filter(a => !selectedActivities.includes(a)) as activity}
                <li on:mousedown={() => toggleActivity(activity)}>{activity}</li>
              {/each}
              <li class="custom-input-item">
                <input
                  type="text"
                  bind:value={newActivityInput}
                  on:keydown={handleActivityKeydown}
                  placeholder="Custom activity..."
                  class="inline-input"
                />
                <button type="button" class="inline-add-btn" on:mousedown={addCustomActivity}>+</button>
              </li>
            </ul>
          {/if}
        </div>
      </div>

      <div class="setting-group">
        <label>PR Size Thresholds</label>
        <p class="help-text">Configure file-count boundaries for PR size categories used in reports.</p>
        <div class="threshold-inputs">
          <div class="threshold-field">
            <label for="pr-small">Small (up to)</label>
            <input id="pr-small" type="number" min="1" bind:value={prSizeSmallMax} />
            <span class="threshold-unit">files</span>
          </div>
          <div class="threshold-field">
            <label for="pr-medium">Medium (up to)</label>
            <input id="pr-medium" type="number" min="1" bind:value={prSizeMediumMax} />
            <span class="threshold-unit">files</span>
          </div>
          <div class="threshold-note">Large = above {prSizeMediumMax} files</div>
        </div>
      </div>

      {#if error}
        <div class="message error-msg" role="alert">{error}</div>
      {/if}
      {#if success}
        <div class="message success-msg" role="status">{success}</div>
      {/if}

      <div class="actions">
        <button type="submit" class="save-btn" disabled={saving || selectedTypes.length === 0}>
          {saving ? 'Saving...' : 'Save Settings'}
        </button>
        <button type="button" class="back-btn" on:click={() => dispatch('back')}>
          Back to Reports
        </button>
      </div>
    </form>
  {/if}
</section>

<style>
  .settings-page {
    background: white;
    border-radius: 10px;
    padding: 2rem;
    box-shadow: 0 1px 4px rgba(46, 58, 61, 0.08), 0 4px 16px rgba(46, 58, 61, 0.06);
  }

  h2 {
    font-family: 'Manrope', 'Inter', system-ui, sans-serif;
    font-size: 1.3rem;
    color: #206473;
    margin-bottom: 1.5rem;
    padding-bottom: 0.75rem;
    border-bottom: 2px solid #e2e8f0;
  }

  .loading-state {
    color: #666;
    padding: 2rem;
    text-align: center;
  }

  .setting-group {
    margin-bottom: 1.75rem;
  }

  .setting-group > label {
    display: block;
    font-size: 0.9rem;
    font-weight: 600;
    color: #333;
    margin-bottom: 0.25rem;
  }

  .help-text {
    font-size: 0.8rem;
    color: #666;
    margin-bottom: 0.5rem;
  }

  input[type="text"] {
    width: 100%;
    padding: 0.6rem 0.75rem;
    border: 1px solid #d0d5dd;
    border-radius: 6px;
    font-size: 0.9rem;
    background: white;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: #206473;
    box-shadow: 0 0 0 3px rgba(32, 100, 115, 0.1);
  }

  .selected-types {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    margin-bottom: 0.5rem;
    min-height: 2rem;
    align-items: center;
  }

  .type-tag {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    background: #e8f4f7;
    color: #206473;
    font-size: 0.8rem;
    font-weight: 500;
    padding: 0.25rem 0.6rem;
    border-radius: 4px;
  }

  .area-tag {
    max-width: 300px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .remove-tag {
    background: none;
    border: none;
    color: #206473;
    font-size: 1rem;
    cursor: pointer;
    padding: 0;
    line-height: 1;
    opacity: 0.6;
  }

  .remove-tag:hover {
    opacity: 1;
  }

  .no-types {
    font-size: 0.8rem;
    color: #999;
    font-style: italic;
  }

  .type-dropdown-wrapper {
    position: relative;
  }

  .dropdown-trigger {
    padding: 0.4rem 0.75rem;
    border: 1px dashed #d0d5dd;
    border-radius: 6px;
    background: #f8fafb;
    color: #666;
    font-size: 0.83rem;
    cursor: pointer;
    transition: all 0.15s;
  }

  .dropdown-trigger:hover {
    border-color: #206473;
    color: #206473;
  }

  .type-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    min-width: 200px;
    max-height: 220px;
    overflow-y: auto;
    background: white;
    border: 1px solid #d0d5dd;
    border-radius: 6px;
    margin-top: 4px;
    list-style: none;
    padding: 0;
    z-index: 100;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .area-dropdown {
    min-width: 320px;
    max-height: 280px;
  }

  .dev-email {
    font-size: 0.72rem;
    color: #888;
  }

  .search-input-item {
    padding: 0.4rem 0.5rem;
    border-bottom: 1px solid #e2e8f0;
    cursor: default;
    position: sticky;
    top: 0;
    background: white;
    z-index: 1;
  }

  .search-input-item:hover {
    background: white !important;
    color: inherit !important;
  }

  .dropdown-search {
    width: 100%;
    padding: 0.35rem 0.5rem;
    border: 1px solid #d0d5dd;
    border-radius: 4px;
    font-size: 0.82rem;
    box-sizing: border-box;
  }

  .threshold-inputs {
    display: flex;
    gap: 1.5rem;
    align-items: flex-end;
    flex-wrap: wrap;
  }

  .threshold-field {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .threshold-field label {
    font-size: 0.78rem;
    color: #666;
    font-weight: 500;
  }

  .threshold-field input {
    width: 80px;
    padding: 0.4rem 0.5rem;
    border: 1px solid #d0d5dd;
    border-radius: 6px;
    font-size: 0.85rem;
  }

  .threshold-unit {
    font-size: 0.72rem;
    color: #888;
  }

  .threshold-note {
    font-size: 0.78rem;
    color: #666;
    align-self: center;
  }

  .dropdown-list-inner {
    list-style: none;
    padding: 0;
    margin: 0;
    max-height: 220px;
    overflow-y: auto;
  }

  .dropdown-list-inner li {
    padding: 0.5rem 0.75rem;
    font-size: 0.85rem;
    cursor: pointer;
    border-bottom: 1px solid #f0f0f0;
  }

  .dropdown-list-inner li:last-child {
    border-bottom: none;
  }

  .dropdown-list-inner li:hover {
    background: #e8f4f7;
    color: #206473;
  }

  .dropdown-list-inner .empty-item {
    color: #999;
    font-style: italic;
    cursor: default;
  }

  .dropdown-list-inner .empty-item:hover {
    background: none;
    color: #999;
  }

  .type-dropdown li {
    padding: 0.5rem 0.75rem;
    font-size: 0.85rem;
    cursor: pointer;
    border-bottom: 1px solid #f0f0f0;
  }

  .type-dropdown li:last-child {
    border-bottom: none;
  }

  .type-dropdown li:hover {
    background: #e8f4f7;
    color: #206473;
  }

  .type-dropdown .empty-item {
    color: #999;
    font-style: italic;
    cursor: default;
  }

  .type-dropdown .empty-item:hover {
    background: none;
    color: #999;
  }

  .custom-input-item {
    display: flex;
    align-items: center;
    gap: 0.3rem;
    padding: 0.4rem 0.5rem;
    border-top: 1px solid #e2e8f0;
    cursor: default;
  }

  .custom-input-item:hover {
    background: none !important;
    color: inherit !important;
  }

  .inline-input {
    flex: 1;
    padding: 0.3rem 0.5rem;
    border: 1px solid #d0d5dd;
    border-radius: 4px;
    font-size: 0.8rem;
    width: auto;
  }

  .inline-add-btn {
    padding: 0.3rem 0.6rem;
    border: none;
    background: #206473;
    color: white;
    border-radius: 4px;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
  }

  .inline-add-btn:hover {
    background: #185364;
  }

  .message {
    padding: 0.75rem 1rem;
    border-radius: 6px;
    font-size: 0.85rem;
    margin-bottom: 1rem;
  }

  .error-msg {
    background: #fde8e8;
    border: 1px solid #f5aca6;
    color: #9b1c1c;
  }

  .success-msg {
    background: #e8f5e9;
    border: 1px solid #a5d6a7;
    color: #2e7d32;
  }

  .actions {
    display: flex;
    gap: 0.75rem;
    align-items: center;
  }

  .save-btn {
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

  .save-btn:hover:not(:disabled) {
    background: #185364;
  }

  .save-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .back-btn {
    padding: 0.6rem 1.5rem;
    background: transparent;
    color: #206473;
    border: 1px solid #206473;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .back-btn:hover {
    background: #e8f4f7;
  }
</style>
