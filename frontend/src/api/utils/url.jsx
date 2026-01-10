import { DISPLAY_BASE_URL } from "../constants/appConfig";

export function buildPrettyShortUrl(code) {
  if (!code) return "";
  return DISPLAY_BASE_URL ? `${DISPLAY_BASE_URL}/${code}` : `/${code}`;
}
