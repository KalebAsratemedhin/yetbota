"use client";

import Link from "next/link";
import { Trophy } from "lucide-react";
import Reveal from "@/components/landing/Reveal";
import SectionShell from "@/components/landing/SectionShell";
import UserAvatar from "@/components/shared/UserAvatar";
import { useListReputationLeadersQuery } from "@/lib/api/socialApi";
import type { UserInfo } from "@/lib/auth/types";
import { useContent } from "@/lib/useContent";
import { cn } from "@/lib/utils";

function formatScore(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, "")}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1).replace(/\.0$/, "")}k`;
  return String(Math.round(n));
}

function ChampionCard({ user, rank }: { user: UserInfo; rank: number }) {
  const display = user.full_name.trim() || user.username;

  return (
    <Link
      href={`/users/${user.user_id}`}
      className={cn(
        "group flex w-44 shrink-0 flex-col items-center rounded-3xl border border-border-subtle bg-bg px-4 py-8 text-center transition-all duration-300 hover:-translate-y-1 hover:border-brand/35 hover:bg-surface-2 sm:w-52 sm:py-10"
      )}
    >
      <div className="relative mb-5">
        <UserAvatar
          user={user}
          size="md"
          className="!h-20 !w-20 !text-base ring-2 ring-brand/30 transition-transform duration-300 group-hover:scale-105"
        />
        <span className="absolute -right-1 -bottom-1 flex h-8 w-8 items-center justify-center rounded-full border border-border-subtle bg-bg text-brand">
          {rank <= 3 ? (
            <Trophy className="h-4 w-4" aria-hidden />
          ) : (
            <span className="text-[11px] font-bold tabular-nums text-fg-muted">
              {rank}
            </span>
          )}
        </span>
      </div>
      <p className="w-full truncate text-base font-bold text-fg group-hover:text-brand">
        {display}
      </p>
      <p className="mt-1 w-full truncate text-xs font-semibold tracking-wide text-brand/90 uppercase">
        {user.reputation_level?.trim() || "Explorer"}
      </p>
      <p className="mt-4 text-sm font-semibold tabular-nums text-fg-muted">
        {formatScore(user.reputation_score ?? 0)} pts
      </p>
    </Link>
  );
}

export default function ChampionsSection() {
  const t = useContent();
  const leadersQ = useListReputationLeadersQuery({ limit: 6 });
  const users = leadersQ.data?.users ?? [];

  return (
    <SectionShell className="bg-bg">
      <Reveal className="mb-8 max-w-2xl">
        <h2 className="mb-2 text-3xl font-extrabold tracking-tight text-fg sm:text-4xl">
          {t.champions.title}
        </h2>
        <p className="text-sm text-fg-muted sm:text-base">
          {t.champions.subtitle}
        </p>
      </Reveal>

      {leadersQ.isLoading ? (
        <div className="flex gap-5 overflow-hidden">
          {Array.from({ length: 4 }, (_, i) => (
            <div
              key={i}
              className="h-56 w-44 shrink-0 animate-pulse rounded-3xl bg-overlay sm:w-52"
              aria-hidden
            />
          ))}
        </div>
      ) : leadersQ.isError ? (
        <p className="max-w-md text-sm text-fg-muted">
          Champions will appear here once the reputation leaderboard is available.
        </p>
      ) : users.length === 0 ? (
        <p className="text-sm text-fg-faint">
          Champions will appear as reputation scores grow.
        </p>
      ) : (
        <div className="yb-scrollbar flex justify-start gap-4 overflow-x-auto pb-2 sm:gap-5">
          {users.map((user, i) => (
            <Reveal key={user.user_id} delay={i * 70} direction="scale">
              <ChampionCard user={user} rank={i + 1} />
            </Reveal>
          ))}
        </div>
      )}
    </SectionShell>
  );
}
