"use client";

import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { LocateFixed, MapPin, Search } from "lucide-react";
import { toast } from "sonner";
import RequireAuth from "@/components/auth/RequireAuth";
import { Navbar } from "@/components/landing";
import PlaceCard from "@/components/places/PlaceCard";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import {
  useLazyListPlacesQuery,
  useListPlacesQuery,
} from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import type { Place } from "@/lib/content/types";
import { cn } from "@/lib/utils";

function PlacesBrowse() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialQ = searchParams.get("q") ?? "";
  const [query, setQuery] = useState(initialQ);
  const debounced = useDebouncedValue(query.trim(), 300);
  const [nearMe, setNearMe] = useState(false);
  const [coords, setCoords] = useState<{ lat: number; lng: number } | null>(
    null
  );
  const [locating, setLocating] = useState(false);

  const listArgs = useMemo(
    () => ({
      limit: 20,
      ...(debounced ? { q: debounced } : {}),
      ...(nearMe && coords
        ? { lat: coords.lat, lng: coords.lng, radius_m: 5000 }
        : {}),
    }),
    [debounced, nearMe, coords]
  );

  const listQ = useListPlacesQuery(listArgs);
  const [fetchMore] = useLazyListPlacesQuery();
  const [extra, setExtra] = useState<Place[]>([]);
  const [cursor, setCursor] = useState<string | null | undefined>(undefined);
  const [loadingMore, setLoadingMore] = useState(false);

  useEffect(() => {
    setExtra([]);
    setCursor(undefined);
  }, [debounced, nearMe, coords?.lat, coords?.lng]);

  useEffect(() => {
    const nextQ = query.trim();
    const currentQ = searchParams.get("q") ?? "";
    if (nextQ === currentQ) return;
    const params = new URLSearchParams();
    if (nextQ) params.set("q", nextQ);
    const qs = params.toString();
    router.replace(qs ? `/places?${qs}` : "/places", { scroll: false });
  }, [query, router, searchParams]);

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
          description: "Allow location access, or search by name instead.",
        });
      },
      { enableHighAccuracy: false, timeout: 12_000 }
    );
  }

  function disableNearMe() {
    setNearMe(false);
  }

  return (
    <div className="mx-auto w-full max-w-3xl px-4 pt-6 pb-16 sm:px-6 sm:pt-8">
      <div className="mb-6 flex items-start gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand/20">
          <MapPin className="h-5 w-5 text-brand" />
        </div>
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-fg">Places</h1>
          <p className="mt-1 text-sm text-fg-muted">
            A directory of real-world spots merged from community posts — tips,
            changes, and what locals know.
          </p>
        </div>
      </div>

      <label className="relative block">
        <span className="sr-only">Search places</span>
        <Search className="pointer-events-none absolute top-1/2 left-3.5 h-4 w-4 -translate-y-1/2 text-fg-faint" />
        <input
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search by place name…"
          className="w-full rounded-xl border border-border-subtle bg-surface-2 py-3 pr-4 pl-10 text-sm text-fg outline-none transition-all placeholder:text-fg-faint focus:border-brand/50 focus:ring-1 focus:ring-brand/20"
        />
      </label>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <button
          type="button"
          onClick={() => (nearMe ? disableNearMe() : enableNearMe())}
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

      <div className="mt-8">
        {listQ.isLoading ? (
          <p className="py-12 text-center text-sm text-fg-faint">
            Loading places…
          </p>
        ) : listQ.isError ? (
          <p className="py-12 text-center text-sm text-fg-muted">
            {getApiErrorMessage(listQ.error)}
          </p>
        ) : places.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-border-subtle px-6 py-14 text-center">
            <p className="text-sm font-semibold text-fg">No places yet</p>
            <p className="mx-auto mt-1 max-w-sm text-xs text-fg-faint">
              {debounced || nearMe
                ? "Try a different name or turn off nearby filter."
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
    </div>
  );
}

export default function PlacesPage() {
  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg text-fg">
        <Navbar />
        <div className="pt-20">
          <Suspense
            fallback={
              <p className="py-20 text-center text-sm text-fg-faint">Loading…</p>
            }
          >
            <PlacesBrowse />
          </Suspense>
        </div>
      </main>
    </RequireAuth>
  );
}
