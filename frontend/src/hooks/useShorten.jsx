import { useState } from "react";
import { apiShorten } from "../api/shortyApi";
import { toRFC3339FromDatetimeLocal } from "../api/utils/dates";

export function useShorten({ onSuccess } = {}) {
  const [url, setUrl] = useState("");
  const [customCode, setCustomCode] = useState("");
  const [wantQR, setWantQR] = useState(true);
  const [expiresAt, setExpiresAt] = useState("");

  const [result, setResult] = useState(null);
  const [errorShorten, setErrorShorten] = useState("");
  const [loadingShorten, setLoadingShorten] = useState(false);

  const clearForm = () => {
    setUrl("");
    setCustomCode("");
    setExpiresAt("");
    setWantQR(true);
    setResult(null);
    setErrorShorten("");
  };

  const handleShorten = async (e) => {
    e.preventDefault();
    setErrorShorten("");
    setResult(null);
    setLoadingShorten(true);

    try {
      const body = {
        url,
        custom_code: customCode || undefined,
        want_qr: wantQR,
        expires_at: expiresAt
          ? toRFC3339FromDatetimeLocal(expiresAt)
          : undefined,
      };

      const data = await apiShorten(body);
      setResult(data);
      onSuccess?.(data);
    } catch (err) {
      setErrorShorten(err.message || "Failed to shorten URL");
    } finally {
      setLoadingShorten(false);
    }
  };

  return {
    // form state
    url,
    setUrl,
    customCode,
    setCustomCode,
    wantQR,
    setWantQR,
    expiresAt,
    setExpiresAt,

    // result state
    result,
    setResult,

    // status
    errorShorten,
    setErrorShorten,
    loadingShorten,

    // actions
    handleShorten,
    clearForm,
  };
}
