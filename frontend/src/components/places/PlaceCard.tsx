"use client";

import Link from "next/link";
import { MapPin } from "lucide-react";
import { formatCoord } from "@/lib/content/helpers";
import type { Place } from "@/lib/content/types";

export default function PlaceCard({ place }: { place: Place }) {
  const summary = place.summary.trim();
  return (
    <li>
      <Link
        href={`/places/${place.id}`}
        className="block rounded-2xl border border-border-subtle bg-surface px-4 py-4 transition-colors hover:border-brand/30 hover:bg-surface-2 sm:px-5"
      >
        <div className="flex items-start justify-between gap-3">
          <h2 className="min-w-0 text-sm font-semibold text-fg sm:text-[15px]">
            {place.name}
          </h2>
          <span className="shrink-0 rounded-full bg-overlay px-2 py-0.5 text-[10px] font-bold tabular-nums text-fg-muted">
            {place.post_count} {place.post_count === 1 ? "post" : "posts"}
          </span>
        </div>
        {summary ? (
          <p className="mt-2 line-clamp-3 text-sm leading-relaxed text-fg-muted">
            {summary}
          </p>
        ) : (
          <p className="mt-2 text-sm text-fg-faint">
            Summary is still being written from community posts.
          </p>
        )}
        <p className="mt-3 inline-flex items-center gap-1.5 text-xs text-fg-faint">
          <MapPin className="h-3.5 w-3.5 text-brand" />
          {formatCoord(place.lat, place.lng)}
        </p>
      </Link>
    </li>
  );
}
