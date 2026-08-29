"use client";

import Image from "next/image";
import { ShieldCheck } from "lucide-react";
import { useContent } from "@/lib/useContent";
import Reveal from "@/components/landing/Reveal";
import CountUp from "@/components/landing/CountUp";

export default function AccuracySection() {
  const t = useContent();

  return (
    <section className="bg-bg px-6 py-16">
      <div className="mx-auto grid max-w-7xl items-center gap-12 md:grid-cols-2">
        <Reveal direction="left" className="relative">
          <div className="relative h-80 w-full overflow-hidden rounded-3xl border border-border-subtle md:h-96">
            <Image
              src="/images/yetbota-bg-2.jpeg"
              alt="Crowdsourced accuracy"
              fill
              className="object-cover"
              sizes="(max-width: 768px) 100vw, 50vw"
            />
          </div>
          <div className="absolute bottom-4 left-4 flex items-center gap-3 rounded-2xl border border-border-subtle bg-bg/95 px-4 py-3">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-brand">
              <ShieldCheck className="h-4 w-4 text-black" />
            </div>
            <div>
              <p className="text-sm font-semibold text-fg">{t.accuracy.verifiedBy}</p>
              <p className="text-xs text-fg-faint">{t.accuracy.lastChecked}</p>
            </div>
          </div>
        </Reveal>

        <Reveal direction="right" delay={150}>
          <p className="mb-4 text-xs font-bold tracking-widest text-brand uppercase">
            {t.accuracy.badge}
          </p>
          <h2 className="mb-6 text-4xl leading-tight font-extrabold text-fg md:text-5xl">
            {t.accuracy.title}
          </h2>
          <p className="mb-8 text-base leading-relaxed text-fg-muted">{t.accuracy.description}</p>

          <div className="grid grid-cols-2 gap-4">
            <div className="rounded-2xl border border-border-subtle bg-bg p-5 transition-all duration-300 hover:-translate-y-1 hover:border-brand/40 hover:shadow-lg hover:shadow-brand/10">
              <CountUp
                value={t.accuracy.stats.verifiedSpots}
                className="mb-1 block text-3xl font-extrabold text-brand"
              />
              <p className="text-xs font-semibold tracking-wider text-fg-faint uppercase">
                {t.accuracy.stats.verifiedSpotsLabel}
              </p>
            </div>
            <div className="rounded-2xl border border-border-subtle bg-bg p-5 transition-all duration-300 hover:-translate-y-1 hover:border-brand/40 hover:shadow-lg hover:shadow-brand/10">
              <CountUp
                value={t.accuracy.stats.communityLed}
                className="mb-1 block text-3xl font-extrabold text-brand"
              />
              <p className="text-xs font-semibold tracking-wider text-fg-faint uppercase">
                {t.accuracy.stats.communityLedLabel}
              </p>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
