"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import ExploreTabs, { type ExploreTab } from "@/components/feed/ExploreTabs";
import FeedShell from "@/components/feed/FeedShell";
import HomeFeed from "@/components/feed/HomeFeed";
import PlacesDirectory from "@/components/places/PlacesDirectory";

function parseTab(raw: string | null): ExploreTab {
  if (raw === "directory" || raw === "places") return "directory";
  return "posts";
}

export default function ExploreHome() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [tab, setTab] = useState<ExploreTab>(() =>
    parseTab(searchParams.get("tab"))
  );

  useEffect(() => {
    setTab(parseTab(searchParams.get("tab")));
  }, [searchParams]);

  function onTabChange(next: ExploreTab) {
    setTab(next);
    const params = new URLSearchParams(searchParams.toString());
    if (next === "posts") params.delete("tab");
    else params.set("tab", "directory");
    const qs = params.toString();
    router.replace(qs ? `/?${qs}` : "/", { scroll: false });
  }

  return (
    <FeedShell
      railTop={
        <ExploreTabs tab={tab} onTabChange={onTabChange} variant="rail" />
      }
    >
      {/* Sidebar is xl+ only — keep a horizontal switcher on smaller screens. */}
      <div className="mx-auto w-full max-w-xl px-3 sm:px-4 xl:hidden">
        <ExploreTabs tab={tab} onTabChange={onTabChange} />
      </div>
      {tab === "posts" ? <HomeFeed bare /> : <PlacesDirectory />}
    </FeedShell>
  );
}
