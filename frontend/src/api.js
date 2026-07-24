const BASE_URL = '/api';

let _token = null;

export function setToken(token) {
  _token = token;
}

async function request(url) {
  const headers = {};
  if (_token) headers['Authorization'] = 'Bearer ' + _token;

  const response = await fetch(url, { headers });
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.error || `Request failed (${response.status})`);
  }
  return response.json();
}

async function requestWithBody(method, url, body) {
  const headers = { 'Content-Type': 'application/json' };
  if (_token) headers['Authorization'] = 'Bearer ' + _token;

  const response = await fetch(url, {
    method,
    headers,
    body: JSON.stringify(body),
  });
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

// Expose an authenticated fetch for components that need raw response handling
export function getAuthHeaders() {
  const headers = {};
  if (_token) headers['Authorization'] = 'Bearer ' + _token;
  return headers;
}

export async function fetchConfig() {
  // Config endpoint is unauthenticated — don't send token
  const response = await fetch(`${BASE_URL}/config`);
  if (!response.ok) throw new Error('Failed to fetch app config');
  return response.json();
}

export async function fetchMe() {
  return request(`${BASE_URL}/me`);
}
