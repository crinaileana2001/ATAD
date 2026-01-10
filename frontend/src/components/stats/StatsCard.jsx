import React from "react";

export default function StatsCard({
  statsCode,
  setStatsCode,
  stats,
  errorStats,
  loadingStats,
  onSubmit,
}) {
  return (
    <div className="card">
      <div className="cardHeader">
        <h2>Statistics</h2>
        <span className="muted">Enter a code or click a row above</span>
      </div>

      <form onSubmit={onSubmit} className="form">
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
            <button className="btn primary" type="submit" disabled={loadingStats}>
              {loadingStats ? "Loading..." : "View stats"}
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
            <a className="link mutedLink" href={stats.original} target="_blank" rel="noreferrer">
              {stats.original}
            </a>
          </div>

          <h3 style={{ marginTop: 16 }}>Clicks by country</h3>
          {!stats.countries || Object.keys(stats.countries).length === 0 ? (
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
  );
}
