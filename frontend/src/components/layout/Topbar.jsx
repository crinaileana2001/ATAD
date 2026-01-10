import React from "react";

export default function Topbar() {
  return (
    <header className="topbar">
      <div>
        <h1 className="title">URL Shortener</h1>
        <p className="subtitle">
          Create short links, track clicks, view analytics — simple & clean.
        </p>
      </div>
      <div className="chip">Backend: /api</div>
    </header>
  );
}
