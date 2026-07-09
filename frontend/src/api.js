const BASE_URL = '/api';

async function request(url) {
  const response = await fetch(url);
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.error || `Request failed (${response.status})`);
  }
  return response.json();
}

export async function fetchDevelopers() {
  const data = await request(`${BASE_URL}/developers`);
  return data.developers || [];
}

export async function fetchReport(developer, from, to) {
  const params = new URLSearchParams({ developer, from, to });
  return request(`${BASE_URL}/report?${params}`);
}

export async function fetchWorkItems(developer, from, to) {
  const params = new URLSearchParams({ developer, from, to });
  return request(`${BASE_URL}/workitems?${params}`);
}

export async function healthCheck() {
  return request(`${BASE_URL}/health`);
}
