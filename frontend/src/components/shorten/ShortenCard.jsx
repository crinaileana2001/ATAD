import React from "react";

export default function ShortenCard({
  url,
  setUrl,
  customCode,
  setCustomCode,
  expiresAt,
  setExpiresAt,
  wantQR,
  setWantQR,
  errorShorten,
  loadingShorten,
  onSubmit,
  onClear,
}) {
  return (
    <div className="card">
      <div className="cardHeader">
        <h2>Create short URL</h2>
        <span className="muted">6–8 chars • collision safe • analytics</span>
      </div>

      <form onSubmit={onSubmit} className="form">
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

        {errorShorten && <div className="alert error">{errorShorten}</div>}

        <div className="actions">
          <button className="btn primary" type="submit" disabled={loadingShorten}>
            {loadingShorten ? "Working..." : "Shorten"}
          </button>
          <button className="btn ghost" type="button" onClick={onClear}>
            Clear
          </button>
        </div>
      </form>
    </div>
  );
}
