"use client";

import { Trophy, Award } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { useContent } from "@/lib/useContent";
import Reveal from "@/components/landing/Reveal";
import AnimatedBar from "@/components/landing/AnimatedBar";

export default function GamificationSection() {
  const t = useContent();

  return (
    <section className="bg-stone-100 px-6 py-16 dark:bg-surface">
      <div className="mx-auto grid max-w-7xl items-center gap-12 md:grid-cols-2">
        <Reveal direction="left">
          <p className="mb-4 text-xs font-bold tracking-widest text-brand uppercase">
            {t.gamification.badge}
          </p>
          <h2 className="mb-6 text-4xl leading-tight font-extrabold text-fg md:text-5xl">
            {t.gamification.title}
          </h2>
          <p className="mb-8 text-base leading-relaxed text-fg-muted">
            {t.gamification.description}
          </p>

          <div className="flex flex-wrap gap-2">
            {t.gamification.badges.map((b) => (
              <Badge
                key={b}
                className="cursor-default rounded-full border border-border-subtle bg-surface-2 px-4 py-1.5 text-xs text-fg-muted transition-colors hover:border-brand/40 hover:text-fg"
              >
                {b}
              </Badge>
            ))}
          </div>
        </Reveal>

        <Reveal
          direction="right"
          delay={150}
          className="rounded-3xl border border-border-subtle bg-bg p-6 transition-all duration-300 hover:border-brand/40 hover:shadow-xl hover:shadow-brand/10"
        >
          <div className="grid grid-cols-2 gap-4">
            <div className="group flex flex-col items-center rounded-2xl bg-bg p-5 text-center">
              <Trophy className="mb-3 h-8 w-8 text-brand transition-transform duration-300 group-hover:scale-125 group-hover:-rotate-6" />
              <p className="mb-1 text-base font-bold text-fg">{t.gamification.levels.current}</p>
              <p className="text-sm text-fg-faint">{t.gamification.levels.currentXp}</p>
            </div>

            <div className="flex flex-col items-center rounded-2xl bg-bg p-5 text-center">
              <Award className="mb-3 h-8 w-8 text-fg-faint" />
              <p className="mb-3 text-base font-bold text-fg">{t.gamification.levels.next}</p>
              <AnimatedBar value={72} />
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
