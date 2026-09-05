/** Google Maps / Maps share-link helpers. */

export type MapsLinkCoords = {
  lat: number;
  lng: number;
  /** Best-effort place name from the URL path, if present. */
  label?: string;
};

const MAPS_HOST =
  /(?:^|\.)(?:google\.[^/]+|goo\.gl|maps\.app\.goo\.gl)$/i;

export function looksLikeGoogleMapsLink(raw: string): boolean {
  const text = raw.trim();
  if (!/^https?:\/\//i.test(text)) return false;
  try {
    const u = new URL(text);
    if (MAPS_HOST.test(u.hostname)) return true;
    if (/google\./i.test(u.hostname) && u.pathname.includes("/maps")) return true;
    return false;
  } catch {
    return false;
  }
}

function finiteCoord(n: number, kind: "lat" | "lng"): number | null {
  if (!Number.isFinite(n)) return null;
  if (kind === "lat" && (n < -90 || n > 90)) return null;
  if (kind === "lng" && (n < -180 || n > 180)) return null;
  return n;
}

function pair(latRaw: string, lngRaw: string): MapsLinkCoords | null {
  const lat = finiteCoord(Number(latRaw), "lat");
  const lng = finiteCoord(Number(lngRaw), "lng");
  if (lat == null || lng == null) return null;
  return { lat, lng };
}

function placeLabelFromPath(pathname: string): string | undefined {
  const m = pathname.match(/\/place\/([^/]+)/i);
  if (!m?.[1]) return undefined;
  try {
    return decodeURIComponent(m[1].replace(/\+/g, " ")).replace(/-/g, " ").trim();
  } catch {
    return m[1].replace(/-/g, " ").trim();
  }
}

/**
 * Extract lat/lng from a (possibly expanded) Google Maps URL.
 * Does not follow redirects — pass the final URL for short links.
 */
export function parseGoogleMapsUrl(raw: string): MapsLinkCoords | null {
  let url: URL;
  try {
    url = new URL(raw.trim());
  } catch {
    return null;
  }

  const label = placeLabelFromPath(url.pathname);
  const href = url.href;

  // /@lat,lng,zoom or /@lat,lngz
  const at = href.match(/@(-?\d+(?:\.\d+)?),(-?\d+(?:\.\d+)?)/);
  if (at) {
    const coords = pair(at[1]!, at[2]!);
    if (coords) return { ...coords, ...(label ? { label } : {}) };
  }

  // !3dLAT!4dLNG (place marker in data blob)
  const bang = href.match(/!3d(-?\d+(?:\.\d+)?)!4d(-?\d+(?:\.\d+)?)/);
  if (bang) {
    const coords = pair(bang[1]!, bang[2]!);
    if (coords) return { ...coords, ...(label ? { label } : {}) };
  }

  const q =
    url.searchParams.get("q") ||
    url.searchParams.get("query") ||
    url.searchParams.get("ll") ||
    url.searchParams.get("center") ||
    url.searchParams.get("destination");

  if (q) {
    const m = q.trim().match(/^(-?\d+(?:\.\d+)?)\s*,\s*(-?\d+(?:\.\d+)?)/);
    if (m) {
      const coords = pair(m[1]!, m[2]!);
      if (coords) return { ...coords, ...(label ? { label } : {}) };
    }
  }

  // /maps/search/lat,lng
  const searchPath = url.pathname.match(
    /\/maps\/search\/\s*(-?\d+(?:\.\d+)?)\s*,\s*(-?\d+(?:\.\d+)?)/i
  );
  if (searchPath) {
    const coords = pair(searchPath[1]!, searchPath[2]!);
    if (coords) return { ...coords, ...(label ? { label } : {}) };
  }

  return null;
}

export function isShortGoogleMapsLink(raw: string): boolean {
  try {
    const u = new URL(raw.trim());
    return (
      u.hostname === "maps.app.goo.gl" ||
      u.hostname === "goo.gl" ||
      u.hostname.endsWith(".app.goo.gl")
    );
  } catch {
    return false;
  }
}
