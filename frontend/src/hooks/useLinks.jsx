import { useCallback, useEffect, useState } from "react";
import { apiListUrls } from "../api/shortyApi";

export function useLinks({ autoLoad = true } = {}) {
  const [allLinks, setAllLinks] = useState([]);
  const [errorLinks, setErrorLinks] = useState("");
  const [linksLoadedOnce, setLinksLoadedOnce] = useState(false);
  const [loadingLinks, setLoadingLinks] = useState(false);

  const loadLinks = useCallback(async () => {
    setErrorLinks("");
    setLoadingLinks(true);
    try {
      const data = await apiListUrls();
      setAllLinks(data);
      setLinksLoadedOnce(true);
    } catch (err) {
      setErrorLinks(err.message || "Failed to load links");
    } finally {
      setLoadingLinks(false);
    }
  }, []);

  useEffect(() => {
    if (autoLoad) loadLinks();
  }, [autoLoad, loadLinks]);

  return {
    allLinks,
    errorLinks,
    linksLoadedOnce,
    loadingLinks,
    loadLinks,
    setAllLinks,
  };
}
