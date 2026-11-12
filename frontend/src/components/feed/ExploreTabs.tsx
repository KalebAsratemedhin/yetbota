"use client";

import { Hash, MapPin } from "lucide-react";
import { cn } from "@/lib/utils";

export type ExploreTab = "posts" | "directory";

export default function ExploreTabs({
  tab,
  onTabChange,
  variant = "bar",
}: {
  tab: ExploreTab;
  onTabChange: (tab: ExploreTab) => void;
  /** `rail` = stacked sidebar control; `bar` = horizontal strip (mobile). */
  variant?: "bar" | "rail";
}) {
  const items = [
    { id: "posts" as const, label: "Posts", short: "Posts", icon: Hash },
    {
      id: "directory" as const,
      label: "Place Directory",
      short: "Places",
      icon: MapPin,
    },
  ] as const;

  if (variant === "rail") {
    return (
      <div className="flex flex-col gap-1 rounded-xl bg-overlay p-1">
        {items.map((t) => {
          const Icon = t.icon;
          return (
            <button
              key={t.id}
              type="button"
              onClick={() => onTabChange(t.id)}
              className={cn(
                "inline-flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm font-semibold transition-colors",
                tab === t.id
                  ? "bg-surface text-fg shadow-sm"
                  : "text-fg-muted hover:text-fg"
              )}
            >
              <Icon className="h-3.5 w-3.5 shrink-0" />
              {t.short}
            </button>
          );
        })}
      </div>
    );
  }

  return (
    <div className="mb-5 flex gap-1 rounded-xl bg-overlay p-1">
      {items.map((t) => {
        const Icon = t.icon;
        return (
          <button
            key={t.id}
            type="button"
            onClick={() => onTabChange(t.id)}
            className={cn(
              "inline-flex flex-1 items-center justify-center gap-1.5 rounded-lg px-3 py-2 text-sm font-semibold transition-colors",
              tab === t.id
                ? "bg-surface text-fg shadow-sm"
                : "text-fg-muted hover:text-fg"
            )}
          >
            <Icon className="h-3.5 w-3.5" />
            {t.label}
          </button>
        );
      })}
    </div>
  );
}
