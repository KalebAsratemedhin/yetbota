"use client";

import Link from "next/link";
import { BadgeCheck, Compass, Footprints, MapPinned, Star, Trophy } from "lucide-react";
import Reveal from "@/components/landing/Reveal";
import SectionShell from "@/components/landing/SectionShell";
import { useContent } from "@/lib/useContent";

const LEVELS = [
  { name: "Explorer", min: 0, icon: Footprints },
  { name: "Contributor", min: 100, icon: BadgeCheck },
  { name: "Local Guide", min: 400, icon: Compass },
  { name: "Pathfinder", min: 1000, icon: MapPinned },
] as const;

const MILESTONES = [
  { name: "Newcomer", icon: Footprints },
  { name: "Legendary explorer", icon: Trophy },
  { name: "Pathfinder", icon: Star },
] as const;

export default function GamificationSection() {
  const t = useContent();

  return (
    <SectionShell className="min-h-[72vh] bg-surface py-28 sm:min-h-[78vh] sm:py-36">
      <div className="grid items-center gap-8 lg:grid-cols-2 lg:gap-12">
        <Reveal direction="left">
          <p className="mb-2 text-[11px] font-bold tracking-[0.18em] text-brand uppercase">
            {t.gamification.badge}
          </p>
          <h2 className="mb-4 text-3xl font-extrabold tracking-tight text-fg sm:text-4xl">
            {t.gamification.title}
          </h2>
          <p className="mb-6 max-w-lg text-sm leading-relaxed text-fg-muted sm:text-base">
            {t.gamification.description}
          </p>
          <Link
            href="/signup"
            className="inline-flex h-11 items-center rounded-xl bg-brand px-5 text-sm font-bold text-black transition-colors hover:bg-brand-dark"
          >
            Start earning reputation
          </Link>
        </Reveal>

        <Reveal direction="right" delay={120} className="space-y-6">
          <div>
            <p className="mb-3 text-xs font-semibold tracking-wide text-fg-faint uppercase">
              Reputation levels
            </p>
            <ul className="grid gap-3 sm:grid-cols-2">
              {LEVELS.map((level) => {
                const Icon = level.icon;
                return (
                  <li
                    key={level.name}
                    className="flex items-center gap-3 rounded-2xl border border-border-subtle bg-bg px-3 py-3"
                  >
                    <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-brand/12 text-brand">
                      <Icon className="h-4 w-4" aria-hidden />
                    </span>
                    <div>
                      <p className="text-sm font-semibold text-fg">{level.name}</p>
                      <p className="text-[11px] text-fg-faint">
                        From {level.min.toLocaleString()} pts
                      </p>
                    </div>
                  </li>
                );
              })}
            </ul>
          </div>

          <div>
            <p className="mb-3 text-xs font-semibold tracking-wide text-fg-faint uppercase">
              Score milestone badges
            </p>
            <div className="flex flex-wrap gap-3">
              {MILESTONES.map((b) => {
                const Icon = b.icon;
                return (
                  <span
                    key={b.name}
                    className="inline-flex items-center gap-2 rounded-full border border-border-subtle bg-bg px-4 py-2 text-sm font-medium text-fg-muted"
                  >
                    <Icon className="h-4 w-4 text-brand" aria-hidden />
                    {b.name}
                  </span>
                );
              })}
            </div>
          </div>
        </Reveal>
      </div>
    </SectionShell>
  );
}
