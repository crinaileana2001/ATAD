import React, { useState } from "react";
import "./App.css";

function App() {
  const [url, setUrl] = useState("");
  const [customCode, setCustomCode] = useState("");
  const [wantQR, setWantQR] = useState(true);
  const [expiresAt, setExpiresAt] = useState("");
  const [result, setResult] = useState(null);
  const [statsCode, setStatsCode] = useState("");
  const [stats, setStats] = useState(null);
  const [errorShorten, setErrorShorten] = useState("");
  const [errorStats, setErrorStats] = useState("");
  const [allLinks, setAllLinks] = useState([]);
  const [errorLinks, setErrorLinks] = useState("");

  // ----------------------------
  // HANDLE SHORTEN
  // ----------------------------
  const handleShorten = async (e) => {
    e.preventDefault();
    setErrorShorten("");
    setResult(null);

    try {
      const body = {
        url,
        custom_code: customCode || undefined,
        want_qr: wantQR,
        expires_at: expiresAt ? new Date(expiresAt).toISOString() : undefined,
      };

      const res = await fetch("/api/shorten", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || "Failed to shorten URL");
      }

      const data = await res.json();
      setResult(data);
      loadLinks();
      setStatsCode(data.code);
    } catch (err) {
      setErrorShorten(err.message);
    }
  };

  // ----------------------------
  // HANDLE STATS
  // ----------------------------
  const handleStats = async (e) => {
    e.preventDefault();
    setErrorStats("");
    setStats(null);

    try {
      const res = await fetch(
        `/api/urls/${encodeURIComponent(statsCode)}/stats`
      );
      if (!res.ok) {
        throw new Error(`Stats not found (${res.status})`);
      }
      const data = await res.json();
      setStats(data);
      loadLinks();
    } catch (err) {
      setErrorStats(err.message);
    }
  };
  // LOAD LINKS //
  const loadLinks = async () => {
    setErrorLinks("");
    try {
      const res = await fetch("/api/urls");
      if (!res.ok) throw new Error(`Failed to load links (${res.status})`);
      const data = await res.json();
      setAllLinks(data);
    } catch (err) {
      setErrorLinks(err.message);
    }
  };
  return (
    <div className="container">
      <h1>URL Shortener</h1>

      {/* ---------------- FORM SHORTEN ----------------- */}
      <form onSubmit={handleShorten} className="card">
        <h2>Create short URL</h2>

        <label>Long URL</label>
        <input
          type="url"
          required
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://example.com/some/long/link"
        />

        <label>Custom code (optional)</label>
        <input
          value={customCode}
          onChange={(e) => setCustomCode(e.target.value)}
          placeholder="ex: crina2025"
        />

        <label>Expires at (optional)</label>
        <input
          type="datetime-local"
          value={expiresAt}
          onChange={(e) => setExpiresAt(e.target.value)}
        />

        <label className="checkbox-row">
          <input
            type="checkbox"
            checked={wantQR}
            onChange={(e) => setWantQR(e.target.checked)}
          />
          <span>Generate QR Code</span>
        </label>

        {errorShorten && <p className="error">{errorShorten}</p>}

        <button type="submit">Shorten</button>
      </form>

      {/* ------------------- RESULT ---------------------- */}
      {result && (
        <div className="card">
          <h2>Result</h2>
          <p>
            <strong>Short URL:</strong>{" "}
            <a href={result.short_url} target="_blank" rel="noreferrer">
              {result.short_url}
            </a>
          </p>
          <p>
            <strong>Code:</strong> <code>{result.code}</code>
          </p>

          {result.qr_base64 && (
            <>
              <h3>QR Code</h3>
              <img src={result.qr_base64} alt="QR" className="qr" />
            </>
          )}
        </div>
      )}

      {/* ------------------- STATS ----------------------- */}
      <form onSubmit={handleStats} className="card">
        <h2>Statistics</h2>
        <div className="card">
          <h2>All links</h2>

          <button type="button" onClick={loadLinks}>
            Refresh list
          </button>
          {errorLinks && <p className="error">{errorLinks}</p>}

          {allLinks.length === 0 ? (
            <p>No links yet.</p>
          ) : (
            <table style={{ width: "100%", marginTop: 12 }}>
              <thead>
                <tr>
                  <th>Code</th>
                  <th>Original</th>
                  <th>Clicks</th>
                  <th>Unique</th>
                  <th>Expires</th>
                </tr>
              </thead>
              <tbody>
                {allLinks.map((l) => (
                  <tr key={l.code}>
                    <td>
                      <a href={l.short_url} target="_blank" rel="noreferrer">
                        {l.code}
                      </a>
                    </td>
                    <td style={{ maxWidth: 320, wordBreak: "break-all" }}>
                      <a href={l.original} target="_blank" rel="noreferrer">
                        {l.original}
                      </a>
                    </td>
                    <td>{l.clicks}</td>
                    <td>{l.unique_visitors}</td>
                    <td>
                      {l.expires_at
                        ? new Date(l.expires_at).toLocaleString()
                        : "-"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        <label>Short code</label>
        <input
          value={statsCode}
          onChange={(e) => setStatsCode(e.target.value)}
          placeholder="ex: crina2025"
        />

        {errorStats && <p className="error">{errorStats}</p>}

        <button type="submit">View stats</button>
      </form>

      {stats && (
        <div className="card">
          <h2>Stats for {statsCode}</h2>
          <p>
            <strong>Original URL:</strong>{" "}
            <a href={stats.original} target="_blank" rel="noreferrer">
              {stats.original}
            </a>
          </p>
          <p>
            <strong>Total clicks:</strong> {stats.clicks}
          </p>

          <p>
            <strong>Unique visitors:</strong> {stats.unique_visitors}
          </p>
          {/* -------- GEO LOCATION -------- */}
          <h3>Clicks by country</h3>

          {!stats.countries || Object.keys(stats.countries).length === 0 ? (
            <p>No geographic data available yet.</p>
          ) : (
            <ul>
              {Object.entries(stats.countries).map(([country, count]) => (
                <li key={country}>
                  🌍 <strong>{country}</strong>: {count} click
                  {count !== 1 ? "s" : ""}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}

export default App;
