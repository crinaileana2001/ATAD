import React, { useState, useEffect } from "react";
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
  const [linksLoadedOnce, setLinksLoadedOnce] = useState(false);

  // Dacă vrei să afișezi un domeniu „frumos” în loc de ngrok, pune aici:
  // ex: const DISPLAY_BASE_URL = "https://shorty.me";
  const DISPLAY_BASE_URL = "";

  // ----------------------------
  // LOAD LINKS
  // ----------------------------
  const loadLinks = async () => {
    setErrorLinks("");
    try {
      const res = await fetch("/api/urls");
      if (!res.ok) throw new Error(`Failed to load links (${res.status})`);
      const data = await res.json();
      setAllLinks(data);
      setLinksLoadedOnce(true);
    } catch (err) {
      setErrorLinks(err.message);
    }
  };

  // opțional: încarcă lista automat la start (doar o dată)
  useEffect(() => {
    loadLinks();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // ----------------------------
  // LOAD STATS
  // ----------------------------
  const loadStats = async (code) => {
    const c = (code || "").trim();
    if (!c) return;

    setErrorStats("");
    setStats(null);
    setStatsCode(c);

    try {
      const res = await fetch(`/api/urls/${encodeURIComponent(c)}/stats`);
      if (!res.ok) throw new Error(`Stats not found (${res.status})`);
      const data = await res.json();
      setStats(data);
    } catch (err) {
      setErrorStats(err.message);
    }
  };

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
      setStatsCode(data.code);
      await loadLinks();
    } catch (err) {
      setErrorShorten(err.message);
    }
  };

  // ----------------------------
  // HANDLE STATS SUBMIT
  // ----------------------------
  const handleStats = async (e) => {
    e.preventDefault();
    await loadStats(statsCode);
  };

  // helpers UI
  const shortText = result
    ? `${DISPLAY_BASE_URL ? DISPLAY_BASE_URL : ""}/${result.code}`
    : "";

  return (
    <div className="page">
      <header className="topbar">
        <div>
          <h1 className="title">URL Shortener</h1>
          <p className="subtitle">
            Create short links, track clicks, view analytics — simple & clean.
          </p>
        </div>
        <div className="chip">Backend: /api</div>
      </header>

      <main className="grid">
        {/* LEFT COLUMN */}
        <section className="stack">
          {/* CREATE */}
          <div className="card">
            <div className="cardHeader">
              <h2>Create short URL</h2>
              <span className="muted">
                6–8 chars • collision safe • analytics
              </span>
            </div>

            <form onSubmit={handleShorten} className="form">
              <div className="field">
                <label>Long URL</label>
                <input
                  type="url"
                  required
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://example.com/some/long/link"
                />
              </div>

              <div className="row2">
                <div className="field">
                  <label>Custom code (optional)</label>
                  <input
                    value={customCode}
                    onChange={(e) => setCustomCode(e.target.value)}
                    placeholder="ex: crina2025"
                  />
                </div>

                <div className="field">
                  <label>Expires at (optional)</label>
                  <input
                    type="datetime-local"
                    value={expiresAt}
                    onChange={(e) => setExpiresAt(e.target.value)}
                  />
                </div>
              </div>

              <div className="qrRow">
                <label className="qrLabel">
                  <input
                    type="checkbox"
                    checked={wantQR}
                    onChange={(e) => setWantQR(e.target.checked)}
                  />
                  <span>Generate QR code</span>
                </label>
              </div>

              {errorShorten && (
                <div className="alert error">{errorShorten}</div>
              )}

              <div className="actions">
                <button className="btn primary" type="submit">
                  Shorten
                </button>
                <button
                  className="btn ghost"
                  type="button"
                  onClick={() => {
                    setUrl("");
                    setCustomCode("");
                    setExpiresAt("");
                    setWantQR(true);
                    setResult(null);
                    setErrorShorten("");
                  }}
                >
                  Clear
                </button>
              </div>
            </form>
          </div>

          {/* RESULT */}
          {result && (
            <div className="card">
              <div className="cardHeader">
                <h2>Result</h2>
                <span className="muted">Click logs are saved on redirect</span>
              </div>

              <div className="resultBox">
                <div className="resultRow">
                  <span className="label">Short URL</span>
                  <a
                    className="mono link"
                    href={result.short_url}
                    target="_blank"
                    rel="noreferrer"
                  >
                    {DISPLAY_BASE_URL
                      ? `${DISPLAY_BASE_URL}/${result.code}`
                      : `/${result.code}`}
                  </a>
                </div>

                <div className="resultRow">
                  <span className="label">Access link</span>
                  <a
                    className="link"
                    href={result.short_url}
                    target="_blank"
                    rel="noreferrer"
                  >
                    Open link
                  </a>
                  <span className="hint">(dev tunnel / current host)</span>
                </div>

                <div className="actions">
                  <button
                    className="btn secondary"
                    type="button"
                    onClick={async () => {
                      try {
                        await navigator.clipboard.writeText(
                          DISPLAY_BASE_URL
                            ? `${DISPLAY_BASE_URL}/${result.code}`
                            : `/${result.code}`
                        );
                      } catch {}
                    }}
                  >
                    Copy short URL
                  </button>

                  <button
                    className="btn ghost"
                    type="button"
                    onClick={async () => {
                      try {
                        await navigator.clipboard.writeText(result.short_url);
                      } catch {}
                    }}
                  >
                    Copy access link
                  </button>
                </div>
              </div>

              {result.qr_base64 && (
                <div className="qrWrap">
                  <h3>QR Code</h3>
                  <img src={result.qr_base64} alt="QR" className="qr" />
                  <p className="muted" style={{ marginTop: 8 }}>
                    Scan to open the access link.
                  </p>
                </div>
              )}
            </div>
          )}
        </section>

        {/* RIGHT COLUMN */}
        <section className="stack">
          {/* ALL LINKS */}
          <div className="card">
            <div className="cardHeader">
              <h2>All links</h2>
              <div className="rightActions">
                {linksLoadedOnce && (
                  <button
                    className="btn secondary"
                    type="button"
                    onClick={loadLinks}
                  >
                    Refresh
                  </button>
                )}
              </div>
            </div>

            {errorLinks && <div className="alert error">{errorLinks}</div>}

            {allLinks.length === 0 ? (
              <div className="empty">
                <p className="muted">No links yet.</p>
                <p className="muted" style={{ marginTop: 4 }}>
                  Create your first short URL on the left.
                </p>
              </div>
            ) : (
              <div className="tableWrap">
                <table className="table">
                  <thead>
                    <tr>
                      <th>Code</th>
                      <th>Original</th>
                      <th className="num">Clicks</th>
                      <th className="num">Unique</th>
                      <th>Expires</th>
                    </tr>
                  </thead>
                  <tbody>
                    {allLinks.map((l) => (
                      <tr
                        key={l.code}
                        className={statsCode === l.code ? "activeRow" : ""}
                        onClick={() => loadStats(l.code)}
                        title="Click to load stats"
                      >
                        <td className="mono">
                          <a
                            className="link"
                            href={l.short_url}
                            target="_blank"
                            rel="noreferrer"
                            onClick={(e) => e.stopPropagation()}
                          >
                            {l.code}
                          </a>
                        </td>
                        <td className="truncate">
                          <a
                            className="link mutedLink"
                            href={l.original}
                            target="_blank"
                            rel="noreferrer"
                            onClick={(e) => e.stopPropagation()}
                          >
                            {l.original}
                          </a>
                        </td>
                        <td className="num">{l.clicks}</td>
                        <td className="num">{l.unique_visitors}</td>
                        <td className="muted">
                          {l.expires_at
                            ? new Date(l.expires_at).toLocaleString()
                            : "—"}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                <div className="muted" style={{ marginTop: 10 }}>
                  Tip: click a row to load stats automatically.
                </div>
              </div>
            )}
          </div>

          {/* STATS */}
          <div className="card">
            <div className="cardHeader">
              <h2>Statistics</h2>
              <span className="muted">Enter a code or click a row above</span>
            </div>

            <form onSubmit={handleStats} className="form">
              <div className="row2">
                <div className="field">
                  <label>Short code</label>
                  <input
                    value={statsCode}
                    onChange={(e) => setStatsCode(e.target.value)}
                    placeholder="ex: crina2025"
                  />
                </div>

                <div className="field" style={{ alignSelf: "end" }}>
                  <button className="btn primary" type="submit">
                    View stats
                  </button>
                </div>
              </div>

              {errorStats && <div className="alert error">{errorStats}</div>}
            </form>

            {stats && (
              <div className="statsBox">
                <div className="statGrid">
                  <div className="stat">
                    <div className="statLabel">Total clicks</div>
                    <div className="statValue">{stats.clicks}</div>
                  </div>
                  <div className="stat">
                    <div className="statLabel">Unique visitors</div>
                    <div className="statValue">{stats.unique_visitors}</div>
                  </div>
                </div>

                <div className="statsRow">
                  <span className="label">Original URL</span>
                  <a
                    className="link mutedLink"
                    href={stats.original}
                    target="_blank"
                    rel="noreferrer"
                  >
                    {stats.original}
                  </a>
                </div>

                <h3 style={{ marginTop: 16 }}>Clicks by country</h3>
                {!stats.countries ||
                Object.keys(stats.countries).length === 0 ? (
                  <p className="muted">No geographic data available yet.</p>
                ) : (
                  <ul className="countryList">
                    {Object.entries(stats.countries).map(([country, count]) => (
                      <li key={country} className="countryItem">
                        <span>🌍 {country}</span>
                        <span className="pill">
                          {count} click{count !== 1 ? "s" : ""}
                        </span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            )}
          </div>
        </section>
      </main>

      <footer className="footer">
        <span className="muted">
          Built with Go + Chi + GORM + SQLite • React UI
        </span>
      </footer>
    </div>
  );
}

export default App;
