"use client";

import Link from "next/link";
import { ChevronDown } from "lucide-react";
import { useContent } from "@/lib/useContent";
import Reveal from "@/components/landing/Reveal";

export default function HeroSection() {
  const t = useContent();

  return (
    <section className="relative flex min-h-[100dvh] items-end overflow-hidden pb-16 pt-28 sm:items-center sm:pb-0 sm:pt-20">
      <div
        className="absolute inset-0 scale-105 bg-cover bg-center bg-no-repeat motion-safe:animate-[welcome-kenburns_28s_ease-in-out_infinite_alternate]"
        style={{ backgroundImage: "url('/images/yetbota-bg-3.jpeg')" }}
        aria-hidden
      />
      <div
        className="absolute inset-0 bg-gradient-to-t from-black via-black/55 to-black/25"
        aria-hidden
      />

      <div className="relative z-10 mx-auto w-full max-w-4xl px-6 text-center sm:px-8">
        <Reveal>
          <p className="mb-4 text-3xl font-extrabold tracking-tight text-white sm:text-4xl md:text-5xl">
            Yet Bota
          </p>
        </Reveal>

        <Reveal delay={120}>
          <h1 className="mb-5 text-balance text-2xl leading-tight font-bold text-white/95 sm:text-3xl md:text-4xl">
            {t.hero.title}{" "}
            <span className="text-brand">{t.hero.titleHighlight}</span>
          </h1>
        </Reveal>

        <Reveal delay={220}>
          <p className="mx-auto mb-9 max-w-lg text-pretty text-base leading-relaxed text-white/75 sm:text-lg">
            {t.hero.subtitle}
          </p>
        </Reveal>

        <Reveal
          delay={320}
          className="flex flex-col items-stretch justify-center gap-3 sm:flex-row sm:items-center"
        >
          <Link
            href="/"
            className="inline-flex items-center justify-center rounded-xl bg-brand px-8 py-3.5 text-sm font-bold text-black transition-colors hover:bg-brand-dark sm:text-base"
          >
            {t.hero.ctaExplore}
          </Link>
          <Link
            href="/signup"
            className="inline-flex items-center justify-center rounded-xl border border-white/70 bg-white/10 px-8 py-3.5 text-sm font-bold text-white backdrop-blur-sm transition-colors hover:bg-white hover:text-black sm:text-base"
          >
            {t.hero.ctaContribute}
          </Link>
        </Reveal>
      </div>

      <div className="absolute bottom-6 left-1/2 z-10 -translate-x-1/2 motion-safe:animate-bounce">
        <ChevronDown className="h-6 w-6 text-white/55" aria-hidden />
      </div>
    </section>
  );
}
