import App from './App.svelte';
import { initKeycloak } from './lib/keycloak.js';
import { setToken, fetchConfig, fetchMe } from './api.js';

async function start() {
  const config = await fetchConfig();

  let currentUser = null;

  if (!config.localDev) {
    const kc = await initKeycloak(config, setToken);
    setToken(kc.token);
  }

  // Fetch the authenticated user (includes roles)
  currentUser = await fetchMe();

  const app = new App({
    target: document.getElementById('app'),
    props: { currentUser },
  });
}

start().catch(err => {
  console.error('Failed to start app:', err);
  document.getElementById('app').innerHTML = `
    <div style="padding: 2rem; text-align: center; color: #c62828;">
      <h2>Unable to start application</h2>
      <p>${err.message}</p>
    </div>
  `;
});
