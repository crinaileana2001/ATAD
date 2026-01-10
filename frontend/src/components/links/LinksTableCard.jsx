import React from "react";
import { formatLocalDateTime } from "./../../api/utils/dates.jsx";

export default function LinksTableCard({
  allLinks,
  errorLinks,
  linksLoadedOnce,
  loadingLinks,
  statsCode,
  onRefresh,
  onRowClick,
}) {
  return (
    <div className="card">
      <div className="cardHeader">
        <h2>All links</h2>
        <div className="rightActions">
          {linksLoadedOnce && (
            <button
              className="btn secondary"
              type="button"
              onClick={onRefresh}
              disabled={loadingLinks}
            >
              {loadingLinks ? "Refreshing..." : "Refresh"}
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
                  onClick={() => onRowClick(l.code)}
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
                  <td className="muted">{formatLocalDateTime(l.expires_at)}</td>
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
  );
}
