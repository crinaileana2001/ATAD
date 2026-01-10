export function toRFC3339FromDatetimeLocal(value) {
  // value: "YYYY-MM-DDTHH:mm" (datetime-local)
  if (!value) return undefined;
  const iso = new Date(value).toISOString();
  return iso;
}

export function formatLocalDateTime(value) {
  if (!value) return "—";
  try {
    return new Date(value).toLocaleString();
  } catch {
    return "—";
  }
}
