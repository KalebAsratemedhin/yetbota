"use client";

import Link from "next/link";
import { MapPin } from "lucide-react";
import type { Place } from "@/lib/content/types";

export default function PlaceCard({ place }: { place: Place }) {
  const summary = place.summary.trim();
  const mapHref = `https://www.openstreetmap.org/?mlat=${place.lat}&mlon=${place.lng}#map=17/${place.lat}/${place.lng}`;

  return (
    <li>
      <div className="rounded-2xl border border-border-subtle bg-surface px-4 py-4 transition-colors hover:border-brand/30 hover:bg-surface-2 sm:px-5">
        <div className="flex items-start justify-between gap-3">
          <Link
            href={`/places/${place.id}`}
            className="min-w-0 text-sm font-semibold text-fg hover:underline sm:text-[15px]"
          >
            {place.name}
          </Link>
          <div className="flex shrink-0 items-center gap-2">
            <a
              href={mapHref}
              target="_blank"
              rel="noreferrer"
              onClick={(e) => e.stopPropagation()}
              className="inline-flex h-7 w-7 items-center justify-center rounded-full text-brand hover:bg-brand/10"
              aria-label={`Open ${place.name} on map`}
              title="Open map"
            >
              <MapPin className="h-3.5 w-3.5" />
            </a>
            <span className="rounded-full bg-overlay px-2 py-0.5 text-[10px] font-bold tabular-nums text-fg-muted">
              {place.post_count} {place.post_count === 1 ? "post" : "posts"}
            </span>
          </div>
        </div>
        <Link href={`/places/${place.id}`} className="mt-2 block">
          {summary ? (
            <p className="line-clamp-2 text-sm leading-relaxed text-fg-muted">
              {summary}
            </p>
          ) : (
            <p className="text-sm text-fg-faint">
              Summary is still being written from community posts.
            </p>
          )}
        </Link>
      </div>
    </li>
  );
}
