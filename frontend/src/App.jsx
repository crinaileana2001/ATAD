import React from "react";
import "./App.css";

import Topbar from "./components/layout/Topbar";
import Footer from "./components/layout/Footer";

import ShortenCard from "./components/shorten/ShortenCard";
import ResultCard from "./components/shorten/ResultCard";
import LinksTableCard from "./components/links/LinksTableCard";
import StatsCard from "./components/stats/StatsCard";

import { useLinks } from "./hooks/useLinks";
import { useStats } from "./hooks/useStats";
import { useShorten } from "./hooks/useShorten";

export default function App() {
  const links = useLinks({ autoLoad: true });
  const stats = useStats();

  const shorten = useShorten({
    onSuccess: async (data) => {
      stats.setStatsCode(data.code);
      await links.loadLinks();
    },
  });

  const onSubmitStats = async (e) => {
    e.preventDefault();
    await stats.loadStats(stats.statsCode);
  };

  return (
    <div className="page">
      <Topbar />

      <main className="grid">
        {/* LEFT COLUMN */}
        <section className="stack">
          <ShortenCard
            url={shorten.url}
            setUrl={shorten.setUrl}
            customCode={shorten.customCode}
            setCustomCode={shorten.setCustomCode}
            expiresAt={shorten.expiresAt}
            setExpiresAt={shorten.setExpiresAt}
            wantQR={shorten.wantQR}
            setWantQR={shorten.setWantQR}
            errorShorten={shorten.errorShorten}
            loadingShorten={shorten.loadingShorten}
            onSubmit={shorten.handleShorten}
            onClear={shorten.clearForm}
          />

          <ResultCard result={shorten.result} />
        </section>

        {/* RIGHT COLUMN */}
        <section className="stack">
          <LinksTableCard
            allLinks={links.allLinks}
            errorLinks={links.errorLinks}
            linksLoadedOnce={links.linksLoadedOnce}
            loadingLinks={links.loadingLinks}
            statsCode={stats.statsCode}
            onRefresh={links.loadLinks}
            onRowClick={(code) => stats.loadStats(code)}
          />

          <StatsCard
            statsCode={stats.statsCode}
            setStatsCode={stats.setStatsCode}
            stats={stats.stats}
            errorStats={stats.errorStats}
            loadingStats={stats.loadingStats}
            onSubmit={onSubmitStats}
          />
        </section>
      </main>

      <Footer />
    </div>
  );
}
