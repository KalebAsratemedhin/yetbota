"use client";

import type { ElementType } from "react";
import {
  BadgeCheck,
  Compass,
  MapPinned,
  ShieldCheck,
  Timer,
  Zap,
} from "lucide-react";

type BadgeDef = {
  id: string;
  label: string;
  category: "exploration" | "reliability" | "speed";
  icon: ElementType;
  tint: string;
  earned: boolean;
};

/** Exploration / reliability / speed — location knowledge discovery. */
const BADGES: BadgeDef[] = [
  {
    id: "explorer",
    label: "Explorer",
    category: "exploration",
    icon: Compass,
    tint: "bg-emerald-500/15 border-emerald-500/30 text-emerald-400",
    earned: true,
  },
  {
    id: "pathfinder",
    label: "Pathfinder",
    category: "exploration",
    icon: MapPinned,
    tint: "bg-yellow-500/15 border-yellow-500/30 text-yellow-400",
    earned: true,
  },
  {
    id: "trusted",
    label: "Trusted",
    category: "reliability",
    icon: ShieldCheck,
    tint: "bg-blue-500/15 border-blue-500/30 text-blue-400",
    earned: true,
  },
  {
    id: "verified",
    label: "Spotter",
    category: "reliability",
    icon: BadgeCheck,
    tint: "bg-purple-500/15 border-purple-500/30 text-purple-400",
    earned: false,
  },
  {
    id: "first-flag",
    label: "First flag",
    category: "speed",
    icon: Zap,
    tint: "bg-rose-500/15 border-rose-500/30 text-rose-400",
    earned: true,
  },
  {
    id: "rapid",
    label: "Rapid",
    category: "speed",
    icon: Timer,
    tint: "bg-orange-500/15 border-orange-500/30 text-orange-400",
    earned: false,
  },
];

export function ReputationCard() {
  const title = "Explorer";
  const percent = 42;

  return (
    <div className="h-full rounded-2xl border border-border-subtle bg-surface p-4 sm:p-5">
      <div className="mb-2 flex items-center justify-between">
        <p className="text-[9px] font-bold tracking-widest text-fg-faint uppercase">
          Reputation Status
        </p>
        <span className="rounded-md bg-brand px-2 py-1 text-[10px] font-bold text-black">
          — pts
        </span>
      </div>
      <h3 className="mb-3 text-lg font-extrabold text-fg">{title}</h3>
      <div className="mb-2">
        <div className="mb-1 flex items-center justify-between">
          <span className="text-xs font-semibold text-fg-faint">
            Progress to Pathfinder
          </span>
          <span className="text-xs font-bold text-brand">{percent}%</span>
        </div>
        <div className="h-2.5 w-full overflow-hidden rounded-full bg-surface-3">
          <div
            className="h-2.5 rounded-full bg-brand shadow-[0px_0px_8px_0px_rgba(34,197,94,0.5)] transition-all"
            style={{ width: `${percent}%` }}
          />
        </div>
      </div>
      <p className="text-xs leading-relaxed text-fg-faint">
        Score and tiers will connect when content reputation lands. Keep mapping
        and verifying spots.
      </p>
    </div>
  );
}

function BadgeItem({ badge }: { badge: BadgeDef }) {
  const Icon = badge.icon;
  return (
    <div className="flex w-20 flex-col items-center gap-1">
      <div
        className={`flex h-12 w-12 items-center justify-center rounded-2xl border ${badge.tint}`}
      >
        <Icon className="h-5 w-5" />
      </div>
      <span className="text-center text-[10px] leading-tight font-bold tracking-wide text-fg-faint uppercase">
        {badge.label}
      </span>
    </div>
  );
}

export function BadgesCard() {
  const earned = BADGES.filter((b) => b.earned);

  return (
    <div className="h-full rounded-2xl border border-border-subtle bg-surface px-4 py-4 sm:px-6 sm:py-5">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-lg font-semibold text-fg">Earned Badges</h3>
        <span className="text-xs font-bold text-fg-faint">{earned.length} earned</span>
      </div>

      <div className="-mx-2 overflow-x-auto px-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        <div className="flex min-w-max items-start gap-6">
          {earned.map((badge) => (
            <BadgeItem key={badge.id} badge={badge} />
          ))}
        </div>
      </div>
    </div>
  );
}
