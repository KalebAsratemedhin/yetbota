import { NextRequest, NextResponse } from "next/server";
import {
  isShortGoogleMapsLink,
  looksLikeGoogleMapsLink,
  parseGoogleMapsUrl,
} from "@/lib/geo/googleMapsLink";

const UA =
  "Mozilla/5.0 (compatible; YetBota/1.0; +https://yetbota.local) AppleWebKit/537.36";

/**
 * Resolve a Google Maps share URL to coordinates.
 * Follows short-link redirects server-side (browser CORS blocks this).
 */
export async function GET(req: NextRequest) {
  const raw = req.nextUrl.searchParams.get("url")?.trim() ?? "";
  if (!raw) {
    return NextResponse.json({ error: "url required" }, { status: 400 });
  }
  if (!looksLikeGoogleMapsLink(raw)) {
    return NextResponse.json({ error: "Not a Google Maps link" }, { status: 400 });
  }

  try {
    let target = raw;
    if (isShortGoogleMapsLink(raw) || !parseGoogleMapsUrl(raw)) {
      const res = await fetch(raw, {
        method: "GET",
        redirect: "follow",
        headers: {
          "User-Agent": UA,
          Accept: "text/html,application/xhtml+xml",
        },
        signal: AbortSignal.timeout(12_000),
      });
      target = res.url || raw;
      // Some share pages embed the destination in meta refresh / canonical.
      if (!parseGoogleMapsUrl(target)) {
        const html = await res.text();
        const canonical =
          html.match(
            /<link[^>]+rel=["']canonical["'][^>]+href=["']([^"']+)["']/i
          )?.[1] ||
          html.match(
            /<meta[^>]+property=["']og:url["'][^>]+content=["']([^"']+)["']/i
          )?.[1] ||
          html.match(/https?:\/\/(?:www\.)?google\.[^"'<\s]+\/maps[^"'<\s]*/i)?.[0];
        if (canonical) target = canonical;
      }
    }

    const coords = parseGoogleMapsUrl(target);
    if (!coords) {
      return NextResponse.json(
        { error: "Could not find coordinates in that Maps link" },
        { status: 422 }
      );
    }
    return NextResponse.json(coords);
  } catch {
    return NextResponse.json(
      { error: "Could not resolve Maps link" },
      { status: 502 }
    );
  }
}
