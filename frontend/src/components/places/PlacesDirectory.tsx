"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { LocateFixed } from "lucide-react";
import { toast } from "sonner";
import PlaceCard from "@/components/places/PlaceCard";
import {
  useLazyListPlacesQuery,
  useListPlacesQuery,
} from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import type { Place } from "@/lib/content/types";
import { cn } from "@/lib/utils";

export default function PlacesDirectory() {
  const [nearMe, setNearMe] = useState(false);
  const [coords, setCoords] = useState<{ lat: number; lng: number } | null>(
    null
  );
  const [locating, setLocating] = useState(false);

  const listArgs = useMemo(
    () => ({
      limit: 20,
      ...(nearMe && coords
        ? { lat: coords.lat, lng: coords.lng, radius_m: 5000 }
        : {}),
    }),
    [nearMe, coords]
  );

  const listQ = useListPlacesQuery(listArgs);
  const [fetchMore] = useLazyListPlacesQuery();
  const [extra, setExtra] = useState<Place[]>([]);
  const [cursor, setCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);

  useEffect(() => {
    setExtra([]);
    setCursor(undefined);
  }, [nearMe, coords?.lat, coords?.lng]);

  const first = listQ.data?.places ?? [];
  const places =
    extra.length === 0
      ? first
      : (() => {
          const seen = new Set(first.map((p) => p.id));
          return [...first, ...extra.filter((p) => !seen.has(p.id))];
        })();

  const next =
    nearMe || cursor === null
      ? undefined
      : cursor === undefined
        ? listQ.data?.next_cursor
        : cursor;

  const loadMore = useCallback(async () => {
    if (!next || loadingMore || nearMe) return;
    setLoadingMore(true);
    try {
      const page = await fetchMore({ ...listArgs, cursor: next }).unwrap();
      setExtra((prev) => [...prev, ...page.places]);
      setCursor(page.next_cursor ?? null);
    } finally {
      setLoadingMore(false);
    }
  }, [next, loadingMore, nearMe, fetchMore, listArgs]);

  function enableNearMe() {
    if (!navigator.geolocation) {
      toast.error("Geolocation is not available in this browser");
      return;
    }
    setLocating(true);
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setCoords({ lat: pos.coords.latitude, lng: pos.coords.longitude });
        setNearMe(true);
        setLocating(false);
      },
      () => {
        setLocating(false);
        toast.error("Could not get your location", {
          description: "Allow location access to filter nearby places.",
        });
      },
      { enableHighAccuracy: false, timeout: 12_000 }
    );
  }

  return (
    <div className="mx-auto w-full max-w-xl px-1 pb-16 sm:px-0">
      <div className="mb-5 flex flex-wrap items-center gap-2">
        <button
          type="button"
          onClick={() => (nearMe ? setNearMe(false) : enableNearMe())}
          disabled={locating}
          className={cn(
            "inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-semibold transition-colors disabled:opacity-60",
            nearMe
              ? "border-brand/40 bg-brand/15 text-brand"
              : "border-border-subtle bg-surface text-fg-muted hover:text-fg"
          )}
        >
          <LocateFixed className="h-3.5 w-3.5" />
          {locating ? "Locating…" : nearMe ? "Near me · on" : "Near me"}
        </button>
        {nearMe ? (
          <span className="text-xs text-fg-faint">Within ~5 km</span>
        ) : null}
      </div>

      {listQ.isLoading ? (
        <p className="py-12 text-center text-sm text-fg-faint">Loading places…</p>
      ) : listQ.isError ? (
        <p className="py-12 text-center text-sm text-fg-muted">
          {getApiErrorMessage(listQ.error)}
        </p>
      ) : places.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-border-subtle px-6 py-14 text-center">
          <p className="text-sm font-semibold text-fg">No places yet</p>
          <p className="mx-auto mt-1 max-w-sm text-xs text-fg-faint">
            {nearMe
              ? "Nothing nearby — try turning off the nearby filter."
              : "Places appear after posts are attached by the AI directory."}
          </p>
        </div>
      ) : (
        <ul className="space-y-3">
          {places.map((p) => (
            <PlaceCard key={p.id} place={p} />
          ))}
        </ul>
      )}

      {next ? (
        <div className="mt-6 flex justify-center">
          <button
            type="button"
            disabled={loadingMore}
            onClick={() => void loadMore()}
            className="rounded-full border border-border-subtle bg-surface px-5 py-2 text-sm font-semibold text-fg hover:bg-surface-2 disabled:opacity-60"
          >
            {loadingMore ? "Loading…" : "Load more"}
          </button>
        </div>
      ) : null}
    </div>
  );
}
