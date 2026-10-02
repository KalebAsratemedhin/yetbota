"use client";

import type { ElementType } from "react";
import {
  BadgeCheck,
  Compass,
  Footprints,
  MapPinned,
  Sparkles,
  Star,
  Trophy,
} from "lucide-react";
import type { Badge } from "@/lib/auth/types";

/** Matches identity-service LevelForScore bands (OpenAPI UserInfo). */
const LEVELS = [
  { name: "Explorer", min: 0 },
  { name: "Contributor", min: 100 },
  { name: "Local Guide", min: 400 },
  { name: "Pathfinder", min: 1000 },
] as const;

/** Icons for identity score-milestone badges (migration 00007/00008). */
const BADGE_ICON: Record<string, { icon: ElementType; tint: string }> = {
  level_newcomer: {
    icon: Footprints,
    tint: "bg-emerald-500/15 border-emerald-500/30 text-emerald-400",
  },
  level_regular: {
    icon: Sparkles,
    tint: "bg-sky-500/15 border-sky-500/30 text-sky-400",
  },
  level_contributor: {
    icon: BadgeCheck,
    tint: "bg-yellow-500/15 border-yellow-500/30 text-yellow-400",
  },
  level_local_guide: {
    icon: Compass,
    tint: "bg-blue-500/15 border-blue-500/30 text-blue-400",
  },
  level_explorer: {
    icon: MapPinned,
    tint: "bg-teal-500/15 border-teal-500/30 text-teal-400",
  },
  level_legendary_explorer: {
    icon: Trophy,
    tint: "bg-amber-500/15 border-amber-500/30 text-amber-400",
  },
  level_pathfinder: {
    icon: Star,
    tint: "bg-violet-500/15 border-violet-500/30 text-violet-400",
  },
};

function levelProgress(score: number, level: string) {
  const idx = LEVELS.findIndex((l) => l.name === level);
  const current = LEVELS[idx >= 0 ? idx : 0]!;
  const next = LEVELS[idx >= 0 ? idx + 1 : 1];
  if (!next) {
    return { percent: 100, nextLabel: null as string | null, remaining: 0 };
  }
  const span = next.min - current.min;
  const gained = Math.max(0, score - current.min);
  const percent = Math.min(100, Math.round((gained / span) * 100));
  return {
    percent,
    nextLabel: next.name,
    remaining: Math.max(0, next.min - score),
  };
}

function LevelGlyph({ level }: { level: string }) {
  const className = "h-4 w-4";
  switch (level) {
    case "Pathfinder":
      return <MapPinned className={className} />;
    case "Local Guide":
      return <Compass className={className} />;
    case "Contributor":
      return <BadgeCheck className={className} />;
    default:
      return <Compass className={className} />;
  }
}

export function ReputationCard({
  score,
  level,
}: {
  score: number;
  level: string;
}) {
  const { percent, nextLabel, remaining } = levelProgress(score, level);

  return (
    <div className="h-full rounded-2xl border border-border-subtle bg-surface p-4 sm:p-5">
      <div className="mb-2 flex items-center justify-between">
        <p className="text-[9px] font-bold tracking-widest text-fg-faint uppercase">
          Reputation Status
        </p>
        <span className="rounded-md bg-brand px-2 py-1 text-[10px] font-bold text-black">
          {score.toLocaleString()} pts
        </span>
      </div>
      <div className="mb-3 flex items-center gap-2">
        <span className="inline-flex h-8 w-8 items-center justify-center rounded-xl border border-brand/30 bg-brand/15 text-brand">
          <LevelGlyph level={level || "Explorer"} />
        </span>
        <h3 className="text-lg font-extrabold text-fg">{level || "Explorer"}</h3>
      </div>
      {nextLabel ? (
        <div className="mb-2">
          <div className="mb-1 flex items-center justify-between">
            <span className="text-xs font-semibold text-fg-faint">
              Progress to {nextLabel}
            </span>
            <span className="text-xs font-bold text-brand">{percent}%</span>
          </div>
          <div className="h-2.5 w-full overflow-hidden rounded-full bg-surface-3">
            <div
              className="h-2.5 rounded-full bg-brand shadow-[0px_0px_8px_0px_rgba(34,197,94,0.5)] transition-all"
              style={{ width: `${percent}%` }}
            />
          </div>
          <p className="mt-2 text-xs leading-relaxed text-fg-faint">
            {remaining} more point{remaining === 1 ? "" : "s"} to reach {nextLabel}.
          </p>
        </div>
      ) : (
        <p className="text-xs leading-relaxed text-fg-faint">
          Top tier unlocked — keep mapping and helping others discover places.
        </p>
      )}
    </div>
  );
}

function BadgeItem({ badge }: { badge: Badge }) {
  const meta = BADGE_ICON[badge.id] ?? {
    icon: BadgeCheck,
    tint: "bg-surface-3 border-border-subtle text-fg-muted",
  };
  const Icon = meta.icon;

  return (
    <div className="flex w-24 flex-col items-center gap-1" title={badge.description}>
      <div
        className={`flex h-12 w-12 items-center justify-center rounded-2xl border ${meta.tint}`}
      >
        <Icon className="h-5 w-5" />
      </div>
      <span className="text-center text-[10px] leading-tight font-bold tracking-wide text-fg-faint uppercase">
        {badge.name}
      </span>
    </div>
  );
}

export function BadgesCard({ badges }: { badges: Badge[] }) {
  const earned = badges ?? [];

  return (
    <div className="h-full rounded-2xl border border-border-subtle bg-surface px-4 py-4 sm:px-6 sm:py-5">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-lg font-semibold text-fg">Earned Badges</h3>
        <span className="text-xs font-bold text-fg-faint">{earned.length} earned</span>
      </div>

      {earned.length === 0 ? (
        <p className="text-sm text-fg-faint">
          No badges yet. Earn reputation to unlock Newcomer and beyond.
        </p>
      ) : (
        <div className="-mx-2 overflow-x-auto px-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          <div className="flex min-w-max items-start gap-6">
            {earned.map((badge) => (
              <BadgeItem key={badge.id} badge={badge} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
