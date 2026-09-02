import { NextRequest, NextResponse } from "next/server";

const NOMINATIM = "https://nominatim.openstreetmap.org";
const UA = "YetBota/1.0 (contact: yetbota@example.com)";

export async function GET(req: NextRequest) {
  const lat = req.nextUrl.searchParams.get("lat");
  const lon = req.nextUrl.searchParams.get("lon");
  if (!lat || !lon) {
    return NextResponse.json({ error: "lat and lon required" }, { status: 400 });
  }
  const latN = Number(lat);
  const lonN = Number(lon);
  if (!Number.isFinite(latN) || !Number.isFinite(lonN)) {
    return NextResponse.json({ error: "lat and lon must be numbers" }, { status: 400 });
  }

  try {
    const url = `${NOMINATIM}/reverse?format=jsonv2&lat=${encodeURIComponent(lat)}&lon=${encodeURIComponent(lon)}&zoom=18`;
    const res = await fetch(url, {
      headers: {
        Accept: "application/json",
        "User-Agent": UA,
      },
      cache: "no-store",
    });
    if (!res.ok) {
      return NextResponse.json(
        { error: "Reverse lookup failed", status: res.status },
        { status: 502 }
      );
    }
    const data: unknown = await res.json();
    return NextResponse.json(data);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Reverse lookup error";
    return NextResponse.json({ error: message }, { status: 502 });
  }
}
