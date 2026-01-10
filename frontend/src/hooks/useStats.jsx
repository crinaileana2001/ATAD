import { useCallback, useState } from "react";
import { apiGetStats } from "../api/shortyApi";

export function useStats() {
  const [statsCode, setStatsCode] = useState("");
  const [stats, setStats] = useState(null);
  const [errorStats, setErrorStats] = useState("");
  const [loadingStats, setLoadingStats] = useState(false);

  const loadStats = useCallback(async (code) => {
    const c = (code || "").trim();
    if (!c) return;

    setErrorStats("");
    setStats(null);
    setStatsCode(c);
    setLoadingStats(true);

    try {
      const data = await apiGetStats(c);
      setStats(data);
    } catch (err) {
      setErrorStats(err.message || "Stats not found");
    } finally {
      setLoadingStats(false);
    }
  }, []);

  return {
    statsCode,
    setStatsCode,
    stats,
    setStats,
    errorStats,
    setErrorStats,
    loadingStats,
    loadStats,
  };
}
