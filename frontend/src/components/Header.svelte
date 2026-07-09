<script>
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  function openHelp() {
    dispatch('openhelp');
  }

  function openSettings() {
    dispatch('opensettings');
  }
</script>

<header>
  <div class="header-content">
    <div class="brand">
      <h1>Delivery Pulse</h1>
    </div>
    <div class="header-actions">
      <button class="icon-btn" on:click={openSettings} title="Settings" aria-label="Open settings">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M6.5 1L6.2 2.6C5.8 2.8 5.4 3 5.1 3.3L3.5 2.7L2 5.3L3.3 6.4C3.3 6.6 3.2 6.8 3.2 7C3.2 7.2 3.2 7.4 3.3 7.6L2 8.7L3.5 11.3L5.1 10.7C5.4 11 5.8 11.2 6.2 11.4L6.5 13H9.5L9.8 11.4C10.2 11.2 10.6 11 10.9 10.7L12.5 11.3L14 8.7L12.7 7.6C12.7 7.4 12.8 7.2 12.8 7C12.8 6.8 12.8 6.6 12.7 6.4L14 5.3L12.5 2.7L10.9 3.3C10.6 3 10.2 2.8 9.8 2.6L9.5 1H6.5ZM8 5C9.1 5 10 5.9 10 7C10 8.1 9.1 9 8 9C6.9 9 6 8.1 6 7C6 5.9 6.9 5 8 5Z" fill="currentColor"/>
        </svg>
      </button>
      <button class="icon-btn" on:click={openHelp} title="Metrics Help" aria-label="Open metrics help">?</button>
    </div>
  </div>
</header>

<style>
  header {
    background: #206473;
    color: white;
    padding: 0.6rem 2rem;
    box-shadow: 0 2px 8px rgba(32, 100, 115, 0.25);
  }

  .header-content {
    max-width: 1200px;
    margin: 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  h1 {
    font-family: 'Manrope', 'Inter', system-ui, sans-serif;
    font-size: 1.1rem;
    font-weight: 700;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.6rem;
    height: 1.6rem;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.15);
    color: rgba(255, 255, 255, 0.85);
    border: none;
    font-weight: 700;
    font-size: 0.85rem;
    cursor: pointer;
    transition: background 0.2s;
  }

  .icon-btn:hover {
    background: rgba(255, 255, 255, 0.25);
    color: white;
  }
</style>
