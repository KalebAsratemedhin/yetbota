import { NextRequest, NextResponse } from "next/server";

const PHOTON = "https://photon.komoot.io/api/";
const NOMINATIM = "https://nominatim.openstreetmap.org";
const UA = "YetBota/1.0 (https://yetbota.local)";

/** Addis Ababa metro — soft bias for search. */
const DEFAULT_LAT = 9.0192;
const DEFAULT_LNG = 38.7525;

export type PlaceHit = {
  lat: number;
  lng: number;
  label: string;
  kind?: string;
};

function asFinite(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim() !== "") {
    const n = Number(value);
    return Number.isFinite(n) ? n : null;
  }
  return null;
}

function haversineKm(
  aLat: number,
  aLng: number,
  bLat: number,
  bLng: number
): number {
  const toRad = (d: number) => (d * Math.PI) / 180;
  const dLat = toRad(bLat - aLat);
  const dLng = toRad(bLng - aLng);
  const s =
    Math.sin(dLat / 2) ** 2 +
    Math.cos(toRad(aLat)) * Math.cos(toRad(bLat)) * Math.sin(dLng / 2) ** 2;
  return 6371 * 2 * Math.asin(Math.sqrt(s));
}

function photonLabel(props: Record<string, unknown>): string {
  const parts = [
    props.name,
    props.housenumber && props.street
      ? `${props.housenumber} ${props.street}`
      : props.street,
    props.city ?? props.district ?? props.county,
    props.state,
    props.country,
  ].filter((p): p is string => typeof p === "string" && p.trim().length > 0);
  const unique: string[] = [];
  for (const p of parts) {
    if (!unique.includes(p)) unique.push(p);
  }
  return unique.join(", ") || "Unknown place";
}

function fromPhoton(data: unknown, biasLat: number, biasLng: number): PlaceHit[] {
  if (!data || typeof data !== "object") return [];
  const features = (data as { features?: unknown }).features;
  if (!Array.isArray(features)) return [];

  const hits: (PlaceHit & { dist: number; cc?: string })[] = [];
  for (const f of features) {
    if (!f || typeof f !== "object") continue;
    const feature = f as {
      geometry?: { coordinates?: unknown };
      properties?: Record<string, unknown>;
    };
    const coords = feature.geometry?.coordinates;
    if (!Array.isArray(coords) || coords.length < 2) continue;
    const lng = asFinite(coords[0]);
    const lat = asFinite(coords[1]);
    if (lat == null || lng == null) continue;
    const props = feature.properties ?? {};
    const kind =
      typeof props.osm_value === "string"
        ? props.osm_value
        : typeof props.osm_key === "string"
          ? props.osm_key
          : undefined;
    hits.push({
      lat,
      lng,
      label: photonLabel(props),
      kind,
      dist: haversineKm(biasLat, biasLng, lat, lng),
      cc: typeof props.countrycode === "string" ? props.countrycode : undefined,
    });
  }

  hits.sort((a, b) => a.dist - b.dist);
  const local = hits.filter((h) => h.cc === "ET" || h.dist <= 120);
  const ranked = (local.length > 0 ? local : hits).slice(0, 10);
  return ranked.map(({ lat, lng, label, kind }) => ({ lat, lng, label, kind }));
}

function fromNominatim(data: unknown): PlaceHit[] {
  if (!Array.isArray(data)) return [];
  const out: PlaceHit[] = [];
  for (const row of data) {
    if (!row || typeof row !== "object") continue;
    const r = row as Record<string, unknown>;
    const lat = asFinite(r.lat);
    const lng = asFinite(r.lon ?? r.lng);
    if (lat == null || lng == null) continue;
    const label =
      (typeof r.display_name === "string" && r.display_name) ||
      (typeof r.name === "string" && r.name) ||
      `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
    const kind = typeof r.type === "string" ? r.type : undefined;
    out.push({ lat, lng, label, kind });
  }
  return out;
}

function dedupe(hits: PlaceHit[]): PlaceHit[] {
  const seen = new Set<string>();
  const out: PlaceHit[] = [];
  for (const h of hits) {
    const key = `${h.lat.toFixed(4)},${h.lng.toFixed(4)}`;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push(h);
  }
  return out;
}

async function fetchPhoton(
  q: string,
  lat: number,
  lng: number
): Promise<PlaceHit[]> {
  const url =
    `${PHOTON}?q=${encodeURIComponent(q)}` +
    `&lat=${lat}&lon=${lng}&limit=12&lang=en`;
  const res = await fetch(url, {
    headers: { Accept: "application/json", "User-Agent": UA },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return fromPhoton(await res.json(), lat, lng);
}

async function fetchNominatim(q: string): Promise<PlaceHit[]> {
  const url =
    `${NOMINATIM}/search?format=jsonv2&limit=6&addressdetails=0` +
    `&countrycodes=et&viewbox=38.2,9.5,39.3,8.6&bounded=0` +
    `&q=${encodeURIComponent(q)}`;
  const res = await fetch(url, {
    headers: { Accept: "application/json", "User-Agent": UA },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return fromNominatim(await res.json());
}

/**
 * Place search: Photon (OSM, strong POI recall) first, Nominatim fills gaps.
 * Normalized `{ lat, lng, label, kind? }[]` for the client.
 */
export async function GET(req: NextRequest) {
  const q = req.nextUrl.searchParams.get("q")?.trim() ?? "";
  if (q.length < 2) {
    return NextResponse.json([]);
  }
  const lat =
    asFinite(req.nextUrl.searchParams.get("lat")) ?? DEFAULT_LAT;
  const lng =
    asFinite(req.nextUrl.searchParams.get("lng")) ??
    asFinite(req.nextUrl.searchParams.get("lon")) ??
    DEFAULT_LNG;

  try {
    const photon = await fetchPhoton(q, lat, lng);
    if (photon.length >= 6) {
      return NextResponse.json(photon);
    }
    const nominatim = await fetchNominatim(q);
    return NextResponse.json(dedupe([...photon, ...nominatim]).slice(0, 10));
  } catch {
    return NextResponse.json({ error: "Place lookup failed" }, { status: 502 });
  }
}
