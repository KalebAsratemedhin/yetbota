"use client";

import { useState } from "react";
import Image from "next/image";
import { Bookmark, HelpCircle, MapPin, Plus } from "lucide-react";
import { cn } from "@/lib/utils";

type ContributionTab = "locations" | "qa" | "saved";

const TABS: { id: ContributionTab; label: string }[] = [
  { id: "locations", label: "Locations" },
  { id: "qa", label: "Q&A" },
  { id: "saved", label: "Saved" },
];

const PLACEHOLDER_CARDS = [
  {
    id: "1",
    title: "Tomoca Coffee – Piazza",
    address: "Piazza, Addis Ababa",
    photo: "/images/places/galani.jpg",
  },
  {
    id: "2",
    title: "Mercato Spice Market",
    address: "Addis Ketema",
    photo: "/images/places/mercato.jpg",
  },
  {
    id: "3",
    title: "Friendship Park",
    address: "Addis Ababa",
    photo: "/images/places/friendship.jpg",
  },
];

function EmptyState({ tab }: { tab: ContributionTab }) {
  const copy: Record<ContributionTab, { title: string; body: string }> = {
    locations: {
      title: "No locations yet",
      body: "Places you contribute will show up here.",
    },
    qa: {
      title: "No Q&A yet",
      body: "Questions and answers you post will land in this tab.",
    },
    saved: {
      title: "No saved places",
      body: "Bookmarks and saved threads will appear here.",
    },
  };
  const c = copy[tab];
  return (
    <div className="flex flex-col items-center justify-center gap-2 rounded-xl border border-border-subtle bg-surface px-6 py-12 text-center">
      {tab === "qa" ? (
        <HelpCircle className="h-8 w-8 text-fg-faint" />
      ) : tab === "saved" ? (
        <Bookmark className="h-8 w-8 text-fg-faint" />
      ) : (
        <MapPin className="h-8 w-8 text-fg-faint" />
      )}
      <p className="text-sm font-semibold text-fg">{c.title}</p>
      <p className="max-w-xs text-xs text-fg-faint">{c.body}</p>
      <span className="mt-2 rounded-full bg-overlay px-2.5 py-1 text-[10px] font-semibold tracking-wide text-fg-faint uppercase">
        Coming soon
      </span>
    </div>
  );
}

function PlaceholderCard({
  title,
  address,
  photo,
}: {
  title: string;
  address: string;
  photo: string;
}) {
  return (
    <div className="group flex cursor-default flex-col overflow-hidden rounded-xl border border-border-subtle bg-surface transition-colors hover:border-overlay-strong">
      <div className="relative aspect-[3/2] overflow-hidden">
        <Image
          src={photo}
          alt={title}
          fill
          className="object-cover transition-transform duration-500 group-hover:scale-105"
          sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
        />
        <span className="absolute top-2 left-2 rounded-full bg-black/55 px-2 py-0.5 text-[10px] font-bold tracking-wide text-white uppercase backdrop-blur-sm">
          Preview
        </span>
      </div>
      <div className="p-3">
        <p className="line-clamp-1 text-sm font-semibold text-fg">{title}</p>
        <div className="mt-1 flex items-center gap-1">
          <MapPin className="h-3 w-3 shrink-0 text-fg-faint" />
          <span className="truncate text-[11px] text-fg-muted">{address}</span>
        </div>
        <div className="mt-2 flex items-center justify-between text-[10px] text-fg-faint">
          <span>—</span>
          <span>0 likes</span>
        </div>
      </div>
    </div>
  );
}

export default function ContributionsPanel() {
  const [tab, setTab] = useState<ContributionTab>("locations");

  return (
    <div className="rounded-2xl border border-border-subtle bg-bg p-4">
      <div className="mb-4 flex items-center justify-between gap-3">
        <h3 className="text-lg font-semibold text-fg">Your Contributions</h3>
        <button
          type="button"
          disabled
          title="Coming soon"
          className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-brand text-black opacity-70"
          aria-label="Create post (coming soon)"
        >
          <Plus className="h-5 w-5" strokeWidth={2.5} />
        </button>
      </div>

      <div
        role="tablist"
        aria-label="Contribution filters"
        className="mb-4 inline-flex w-fit rounded-xl bg-overlay p-1"
      >
        {TABS.map((t) => (
          <button
            key={t.id}
            type="button"
            role="tab"
            aria-selected={tab === t.id}
            onClick={() => setTab(t.id)}
            className={cn(
              "rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors sm:text-sm",
              tab === t.id
                ? "bg-surface text-fg shadow-sm"
                : "text-fg-muted hover:text-fg"
            )}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === "locations" ? (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {PLACEHOLDER_CARDS.map((card) => (
            <PlaceholderCard key={card.id} {...card} />
          ))}
        </div>
      ) : (
        <EmptyState tab={tab} />
      )}
    </div>
  );
}
