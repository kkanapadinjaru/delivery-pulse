<script>
  import { createEventDispatcher, onMount } from 'svelte';

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

  let teams = '';
  let selectedTypes = [];
  let loading = true;
  let saving = false;
  let error = '';
  let success = '';
  let showTypeDropdown = false;

  onMount(async () => {
    try {
      const resp = await fetch('/api/settings');
      if (!resp.ok) throw new Error('Failed to load settings');
      const data = await resp.json();
      teams = (data.teams || []).join(', ');
      selectedTypes = data.workItemTypes || ['Bug', 'Task'];
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
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          teams: teamsList,
          workItemTypes: selectedTypes,
        }),
      });

      if (!resp.ok) {
        const data = await resp.json().catch(() => ({}));
        throw new Error(data.error || 'Failed to save settings');
      }

      success = 'Settings saved successfully. Changes will take effect on the next report generation.';
      // Notify parent that settings changed (so developer list can be refreshed)
      dispatch('saved');
    } catch (e) {
      error = e.message;
    } finally {
      saving = false;
    }
  }

  function handleDropdownBlur() {
    setTimeout(() => { showTypeDropdown = false; }, 150);
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
            on:blur={handleDropdownBlur}
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
