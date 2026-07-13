const BASE = '/api';

let token = null;

export function setAuthToken(t) {
  token = t;
}

function headers() {
  const h = { 'Content-Type': 'application/json' };
  if (token) h['Authorization'] = `Bearer ${token}`;
  return h;
}

async function req(url, opts = {}) {
  const res = await fetch(BASE + url, {
    headers: { ...headers(), ...opts.headers },
    ...opts,
  });
  if (res.status === 204) return null;
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
  return data;
}

export function login(email, password) {
  return req('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
}

export function signup(email, password, display_name) {
  return req('/auth/signup', {
    method: 'POST',
    body: JSON.stringify({ email, password, display_name }),
  });
}

export function getMe() {
  return req('/auth/me');
}

export function logout() {
  return req('/auth/logout', { method: 'POST' });
}

export function listHotels(page = 1, location = '', withPrices = true) {
  const params = new URLSearchParams({ page, limit: '20' });
  if (location) params.set('location', location);
  if (withPrices) params.set('with_prices', 'true');
  return req(`/hotels?${params}`);
}

export function getHotel(id) {
  return req(`/hotels/${id}`);
}

export function getHotelReviews(id) {
  return req(`/hotels/${id}/reviews`);
}

export function getHotelPrices(id) {
  return req(`/hotels/${id}/prices`);
}

export function getHotelRatings(id) {
  return req(`/hotels/${id}/ratings`);
}

export function triggerCollect() {
  return req('/collect', { method: 'POST' });
}

export function listSettings() {
  return req('/settings');
}

export function getSetting(key) {
  return req(`/settings/${encodeURIComponent(key)}`);
}

export function upsertSetting(key, value) {
  return req('/settings', {
    method: 'POST',
    body: JSON.stringify({ key, value }),
  });
}

export function updateSetting(key, value) {
  return req(`/settings/${encodeURIComponent(key)}`, {
    method: 'PUT',
    body: JSON.stringify({ value }),
  });
}

export function deleteSetting(key) {
  return req(`/settings/${encodeURIComponent(key)}`, { method: 'DELETE' });
}

export function getVacations(year) {
  const params = year ? `?year=${year}` : '';
  return req(`/vacations${params}`);
}
