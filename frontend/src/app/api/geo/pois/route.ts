import { NextRequest, NextResponse } from "next/server";

const OVERPASS = "https://overpass-api.de/api/interpreter";
const UA = "YetBota/1.0 (https://yetbota.local)";

export type MapPoi = {
  id: string;
  lat: number;
  lng: number;
  name: string;
  kind: string;
};

function asFinite(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim() !== "") {
    const n = Number(value);
    return Number.isFinite(n) ? n : null;
  }
  return null;
}

/**
 * Nearby named POIs for the visible map (shops, amenities, tourism, …).
 * Query is bounded so we never hammer Overpass with a huge bbox.
 */
export async function GET(req: NextRequest) {
  const south = asFinite(req.nextUrl.searchParams.get("south"));
  const west = asFinite(req.nextUrl.searchParams.get("west"));
  const north = asFinite(req.nextUrl.searchParams.get("north"));
  const east = asFinite(req.nextUrl.searchParams.get("east"));

  if (south == null || west == null || north == null || east == null) {
    return NextResponse.json({ error: "bbox required" }, { status: 400 });
  }
  if (south >= north || west >= east) {
    return NextResponse.json({ error: "invalid bbox" }, { status: 400 });
  }

  const latSpan = north - south;
  const lngSpan = east - west;
  // ~0.06° ≈ 6–7 km — refuse huge areas (zoom out)
  if (latSpan > 0.08 || lngSpan > 0.08) {
    return NextResponse.json([]);
  }

  const query = `
[out:json][timeout:12];
(
  nwr["name"]["amenity"](${south},${west},${north},${east});
  nwr["name"]["shop"](${south},${west},${north},${east});
  nwr["name"]["tourism"](${south},${west},${north},${east});
  nwr["name"]["leisure"](${south},${west},${north},${east});
  nwr["name"]["healthcare"](${south},${west},${north},${east});
  nwr["name"]["office"](${south},${west},${north},${east});
  nwr["name"]["craft"](${south},${west},${north},${east});
  nwr["name"]["historic"](${south},${west},${north},${east});
  nwr["name"]["building"="commercial"](${south},${west},${north},${east});
);
out center tags 70;
`.trim();

  try {
    const res = await fetch(OVERPASS, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
        Accept: "application/json",
        "User-Agent": UA,
      },
      body: `data=${encodeURIComponent(query)}`,
      cache: "no-store",
    });
    if (!res.ok) {
      return NextResponse.json({ error: "Overpass failed" }, { status: 502 });
    }

    const data = (await res.json()) as {
      elements?: Array<{
        type: string;
        id: number;
        lat?: number;
        lon?: number;
        center?: { lat?: number; lon?: number };
        tags?: Record<string, string>;
      }>;
    };

    const pois: MapPoi[] = [];
    for (const el of data.elements ?? []) {
      const tags = el.tags ?? {};
      const name = tags.name?.trim();
      if (!name) continue;
      const lat = asFinite(el.lat ?? el.center?.lat);
      const lng = asFinite(el.lon ?? el.center?.lon);
      if (lat == null || lng == null) continue;
      const kind =
        tags.amenity ||
        tags.shop ||
        tags.tourism ||
        tags.leisure ||
        tags.healthcare ||
        tags.office ||
        tags.craft ||
        tags.historic ||
        tags.building ||
        "place";
      pois.push({
        id: `${el.type}/${el.id}`,
        lat,
        lng,
        name,
        kind,
      });
      if (pois.length >= 60) break;
    }

    return NextResponse.json(pois);
  } catch {
    return NextResponse.json({ error: "Overpass failed" }, { status: 502 });
  }
}
