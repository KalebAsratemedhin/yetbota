"use client";

import { Star, Trophy } from "lucide-react";
import { useContent } from "@/lib/useContent";
import Reveal from "@/components/landing/Reveal";

const AVATAR_COLORS = ["bg-purple-700", "bg-teal-700", "bg-indigo-700", "bg-rose-700"];
const ROLE_COLORS = ["text-yellow-400", "text-brand", "text-blue-400", "text-orange-400"];

export default function ChampionsSection() {
  const t = useContent();
  const champions = t.champions.users;
  const slides = [...champions, ...champions];

  return (
    <section className="bg-stone-100 py-16 dark:bg-bg">
      <div className="mx-auto max-w-7xl px-6 lg:px-12">
        <Reveal className="mb-12 text-center">
          <h2 className="mb-3 text-4xl font-extrabold text-fg">{t.champions.title}</h2>
          <p className="text-base text-fg-faint">{t.champions.subtitle}</p>
        </Reveal>

        <div className="group relative mx-auto max-w-4xl overflow-hidden py-6">
          <div className="animate-champions-scroll flex w-max gap-8 group-hover:[animation-play-state:paused]">
            {slides.map((champion, i) => {
              const colorIdx = i % AVATAR_COLORS.length;
              const initials = champion.name
                .split(" ")
                .filter(Boolean)
                .map((n) => n[0])
                .join("")
                .slice(0, 2)
                .toUpperCase();

              return (
                <div
                  key={`${champion.name}-${i}`}
                  className="group/item flex w-40 shrink-0 flex-col items-center px-2 text-center"
                >
                  <div className="relative mb-4">
                    <div
                      className={`flex h-20 w-20 items-center justify-center overflow-hidden rounded-full border-2 border-brand/40 text-xl font-bold text-white transition-all duration-300 group-hover/item:scale-110 group-hover/item:border-brand ${AVATAR_COLORS[colorIdx]}`}
                    >
                      {initials}
                    </div>
                    <div className="absolute -right-1 -bottom-1 flex h-8 w-8 items-center justify-center rounded-full border border-border-subtle bg-bg transition-transform duration-300 group-hover/item:scale-110">
                      <Trophy className="h-4 w-4 text-brand" />
                    </div>
                  </div>
                  <p className="mb-0.5 text-base font-bold text-fg transition-colors group-hover/item:text-brand">
                    {champion.name}
                  </p>
                  <p
                    className={`mb-3 text-xs font-bold tracking-widest uppercase ${ROLE_COLORS[colorIdx]}`}
                  >
                    {champion.role}
                  </p>
                  <div className="inline-flex items-center gap-1.5">
                    <Star className="h-3.5 w-3.5 shrink-0 fill-yellow-400 text-yellow-400" />
                    <span className="text-sm font-semibold text-fg-muted">{champion.points}</span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </section>
  );
}
