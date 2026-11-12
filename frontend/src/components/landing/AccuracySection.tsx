"use client";

import { Camera, MapPinned, Users } from "lucide-react";
import Reveal from "@/components/landing/Reveal";
import SectionShell from "@/components/landing/SectionShell";
import { useContent } from "@/lib/useContent";

const PILLARS = [
  {
    icon: Users,
    title: "Locals first",
    body: "Pins and notes come from people who actually go there — not scraped listings.",
  },
  {
    icon: Camera,
    title: "Media that proves it",
    body: "Drops include photos or video so you can see the place before you go.",
  },
  {
    icon: MapPinned,
    title: "Coordinates that stick",
    body: "Every post carries a real lat/lng so navigation stays precise as the city changes.",
  },
] as const;

export default function AccuracySection() {
  const t = useContent();

  return (
    <SectionShell className="min-h-[72vh] bg-bg py-28 sm:min-h-[78vh] sm:py-36">
      <Reveal className="mb-8 max-w-2xl">
        <p className="mb-2 text-[11px] font-bold tracking-[0.18em] text-brand uppercase">
          {t.accuracy.badge}
        </p>
        <h2 className="mb-3 text-3xl font-extrabold tracking-tight text-fg sm:text-4xl">
          {t.accuracy.title}
        </h2>
        <p className="text-sm leading-relaxed text-fg-muted sm:text-base">
          {t.accuracy.description}
        </p>
      </Reveal>

      <ul className="grid gap-4 sm:grid-cols-3">
        {PILLARS.map((item, i) => {
          const Icon = item.icon;
          return (
            <Reveal key={item.title} delay={i * 60}>
              <li className="rounded-2xl border border-border-subtle bg-surface px-4 py-5">
                <span className="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-brand/12 text-brand">
                  <Icon className="h-5 w-5" aria-hidden />
                </span>
                <p className="text-sm font-semibold text-fg">{item.title}</p>
                <p className="mt-1.5 text-sm leading-relaxed text-fg-faint">
                  {item.body}
                </p>
              </li>
            </Reveal>
          );
        })}
      </ul>
    </SectionShell>
  );
}
