export type GeoHit = {
  lat: number;
  lng: number;
  label: string;
  kind?: string;
};

export type MapPoi = {
  id: string;
  lat: number;
  lng: number;
  name: string;
  kind: string;
};

export type MapView = {
  south: number;
  west: number;
  north: number;
  east: number;
  zoom: number;
  centerLat: number;
  centerLng: number;
};

/** Addis Ababa — default map center for Yet Bota. */
export const DEFAULT_CENTER = { lat: 9.0192, lng: 38.7525 };

/** Zoom at which we load named POIs onto the map. */
export const POI_MIN_ZOOM = 15;

export {
  looksLikeGoogleMapsLink,
  parseGoogleMapsUrl,
  isShortGoogleMapsLink,
  type MapsLinkCoords,
} from "@/lib/geo/googleMapsLink";

export async function resolveGoogleMapsLink(
  url: string
): Promise<{ lat: number; lng: number; label?: string }> {
  const res = await fetch(
    `/api/geo/maps-link?url=${encodeURIComponent(url.trim())}`
  );
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { error?: string } | null;
    throw new Error(body?.error || "Could not resolve Maps link");
  }
  const data = (await res.json()) as {
    lat?: number;
    lng?: number;
    label?: string;
  };
  if (
    typeof data.lat !== "number" ||
    typeof data.lng !== "number" ||
    !Number.isFinite(data.lat) ||
    !Number.isFinite(data.lng)
  ) {
    throw new Error("Could not read coordinates from Maps link");
  }
  return {
    lat: data.lat,
    lng: data.lng,
    ...(data.label ? { label: data.label } : {}),
  };
}

export async function searchPlaces(
  q: string,
  bias?: { lat: number; lng: number }
): Promise<GeoHit[]> {
  const query = q.trim();
  if (query.length < 2) return [];
  const params = new URLSearchParams({ q: query });
  if (bias && Number.isFinite(bias.lat) && Number.isFinite(bias.lng)) {
    params.set("lat", String(bias.lat));
    params.set("lng", String(bias.lng));
  }
  const res = await fetch(`/api/geo/search?${params}`);
  if (!res.ok) throw new Error("Place lookup failed");
  const data: unknown = await res.json();
  if (!Array.isArray(data)) return [];
  return data.filter((row): row is GeoHit => {
    if (!row || typeof row !== "object") return false;
    const h = row as GeoHit;
    return (
      Number.isFinite(h.lat) &&
      Number.isFinite(h.lng) &&
      typeof h.label === "string" &&
      h.label.length > 0
    );
  });
}

export async function reversePlace(
  lat: number,
  lng: number,
  signal?: AbortSignal
): Promise<string> {
  if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
    throw new Error("Invalid coordinates");
  }
  const res = await fetch(
    `/api/geo/reverse?lat=${encodeURIComponent(String(lat))}&lon=${encodeURIComponent(String(lng))}`,
    { signal }
  );
  if (!res.ok) throw new Error("Reverse lookup failed");
  const row = (await res.json()) as { display_name?: string };
  return row.display_name || `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
}

export async function fetchMapPois(view: MapView): Promise<MapPoi[]> {
  if (view.zoom < POI_MIN_ZOOM) return [];
  const params = new URLSearchParams({
    south: String(view.south),
    west: String(view.west),
    north: String(view.north),
    east: String(view.east),
  });
  const res = await fetch(`/api/geo/pois?${params}`);
  if (!res.ok) return [];
  const data: unknown = await res.json();
  if (!Array.isArray(data)) return [];
  return data.filter((row): row is MapPoi => {
    if (!row || typeof row !== "object") return false;
    const p = row as MapPoi;
    return (
      typeof p.id === "string" &&
      typeof p.name === "string" &&
      Number.isFinite(p.lat) &&
      Number.isFinite(p.lng)
    );
  });
}
