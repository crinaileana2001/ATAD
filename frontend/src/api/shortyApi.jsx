import { API_BASE } from "./constants/appConfig";

async function readError(res) {
  const text = await res.text().catch(() => "");
  return text || `Request failed (${res.status})`;
}

export async function apiListUrls() {
  const res = await fetch(`${API_BASE}/api/urls`);
  if (!res.ok) throw new Error(await readError(res));
  return res.json();
}

export async function apiGetStats(code) {
  const c = (code || "").trim();
  const res = await fetch(`${API_BASE}/api/urls/${encodeURIComponent(c)}/stats`);
  if (!res.ok) throw new Error(await readError(res));
  return res.json();
}

export async function apiShorten(body) {
  const res = await fetch(`${API_BASE}/api/shorten`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(await readError(res));
  return res.json();
}
