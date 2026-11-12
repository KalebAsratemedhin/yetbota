"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import SnapFeed from "@/components/feed/SnapFeed";
import FeedShell from "@/components/feed/FeedShell";
import { useLocationPreference } from "@/hooks/use-location-preference";
import { useGetFeedQuery, useLazyGetFeedQuery } from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import type { FeedPage } from "@/lib/content/types";
import {
  consumeLocationPrompt,
  writeLocationPref,
} from "@/lib/locationPrefs";

type GeoState =
  | { status: "pending" }
  | { status: "ready"; lat: number; lng: number }
  | { status: "unavailable" }
  | { status: "skipped" };

function readGeo(): Promise<{ lat: number; lng: number }> {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error("Geolocation is not available in this browser."));
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) =>
        resolve({
          lat: pos.coords.latitude,
          lng: pos.coords.longitude,
        }),
      (err) => {
        reject(
          new Error(
            err.code === err.PERMISSION_DENIED
              ? "Location permission denied"
              : "Could not read location"
          )
        );
      },
      { enableHighAccuracy: true, timeout: 15000, maximumAge: 60_000 }
    );
  });
}

export default function HomeFeed({ bare = false }: { bare?: boolean }) {
  const { enabled: locationEnabled } = useLocationPreference();
  const [geo, setGeo] = useState<GeoState>({ status: "pending" });
  const [geoTick, setGeoTick] = useState(0);
  const [promptAfterAuth] = useState(() => consumeLocationPrompt());

  useEffect(() => {
    if (!locationEnabled) {
      setGeo({ status: "skipped" });
      return;
    }

    let cancelled = false;
    setGeo({ status: "pending" });
    void readGeo()
      .then((coords) => {
        if (!cancelled) setGeo({ status: "ready", ...coords });
      })
      .catch(() => {
        if (!cancelled) setGeo({ status: "unavailable" });
      });
    return () => {
      cancelled = true;
    };
  }, [geoTick, locationEnabled]);

  useEffect(() => {
    if (geo.status !== "unavailable") return;
    if (!promptAfterAuth) return;
    toast.message("Location off", {
      id: "feed-geo-unavailable",
      description: "Showing following / recent posts. Enable location for nearby.",
      duration: 6000,
      action: {
        label: "Enable",
        onClick: () => {
          writeLocationPref("on");
          setGeo({ status: "pending" });
          setGeoTick((n) => n + 1);
        },
      },
    });
  }, [geo.status, promptAfterAuth]);

  const feedArgs = useMemo(
    () =>
      geo.status === "ready"
        ? { lat: geo.lat, lng: geo.lng, limit: 12 }
        : { limit: 12 },
    [geo]
  );

  const feedQ = useGetFeedQuery(feedArgs);
  const [fetchMore] = useLazyGetFeedQuery();
  const lastFeedRef = useRef<FeedPage | null>(null);
  if (feedQ.data) lastFeedRef.current = feedQ.data;
  const feedData = feedQ.data ?? lastFeedRef.current;

  const onLoadMore = useCallback(
    async (cursor: string): Promise<FeedPage> => {
      return fetchMore({
        ...feedArgs,
        cursor,
      }).unwrap();
    },
    [fetchMore, feedArgs]
  );

  if (feedQ.isError && !feedData) {
    return (
      <div className="flex min-h-[calc(100dvh-5rem)] flex-col items-center justify-center gap-3 px-4 text-center">
        <p className="text-sm text-fg-muted">
          {getApiErrorMessage(feedQ.error) || "Could not load feed."}
        </p>
        <button
          type="button"
          onClick={() => void feedQ.refetch()}
          className="text-sm font-semibold text-brand hover:underline"
        >
          Retry
        </button>
      </div>
    );
  }

  if (!feedData) {
    return (
      <div className="flex min-h-[calc(100dvh-5rem)] flex-col items-center justify-center gap-3 px-4">
        <Loader2 className="h-6 w-6 animate-spin text-brand" />
        <p className="text-sm text-fg-muted">Loading your feed…</p>
      </div>
    );
  }

  const feedKey =
    geo.status === "ready"
      ? `${geo.lat.toFixed(3)}:${geo.lng.toFixed(3)}`
      : locationEnabled
        ? "no-geo"
        : "loc-off";

  const feed = (
    <SnapFeed
      key={feedKey}
      embedded
      initialPosts={feedData.posts}
      initialPhase={feedData.phase}
      nextCursor={feedData.next_cursor}
      onLoadMore={onLoadMore}
    />
  );

  if (bare) return feed;
  return <FeedShell>{feed}</FeedShell>;
}
