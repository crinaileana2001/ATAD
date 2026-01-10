import React from "react";
import { copyToClipboard } from "../../api/utils/clipboard";
import { buildPrettyShortUrl } from "../../api/utils/url";

export default function ResultCard({ result }) {
  if (!result) return null;

  const pretty = buildPrettyShortUrl(result.code);

  return (
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
            {pretty}
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
            onClick={() => copyToClipboard(pretty)}
          >
            Copy short URL
          </button>

          <button
            className="btn ghost"
            type="button"
            onClick={() => copyToClipboard(result.short_url)}
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
  );
}
