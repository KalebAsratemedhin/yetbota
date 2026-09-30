"use client";

import Link from "next/link";
import { ChevronDown } from "lucide-react";
import { useContent } from "@/lib/useContent";
import Reveal from "@/components/landing/Reveal";

export default function HeroSection() {
  const t = useContent();

  return (
    <section className="relative flex min-h-screen items-center justify-center overflow-hidden">
      <div
        className="absolute inset-0 bg-cover bg-center bg-no-repeat"
        style={{ backgroundImage: "url('/images/yetbota-bg-3.jpeg')" }}
      />
      <div className="absolute inset-0 bg-black/60" />

      <div className="relative z-10 mx-auto max-w-3xl px-6 pt-20 text-center">
        <Reveal>
          <h1 className="mb-6 text-5xl leading-tight font-extrabold text-white md:text-7xl">
            {t.hero.title}{" "}
            <span className="text-brand">{t.hero.titleHighlight}</span>
          </h1>
        </Reveal>

        <Reveal delay={150}>
          <p className="mx-auto mb-10 max-w-xl text-lg leading-relaxed text-white/80 md:text-xl">
            {t.hero.subtitle}
          </p>
        </Reveal>

        <Reveal
          delay={300}
          className="flex flex-col items-center justify-center gap-4 sm:flex-row"
        >
          <Link
            href="/places"
            className="inline-flex w-full items-center justify-center rounded-xl bg-brand px-8 py-4 text-base font-bold text-black transition-colors hover:bg-brand-dark sm:w-auto"
          >
            {t.hero.ctaExplore}
          </Link>
          <Link
            href="/create"
            className="inline-flex w-full items-center justify-center rounded-xl border-2 border-white bg-white px-8 py-4 text-base font-bold text-black transition-colors duration-200 hover:bg-transparent hover:text-white sm:w-auto"
          >
            {t.hero.ctaContribute}
          </Link>
        </Reveal>
      </div>

      <div className="absolute bottom-8 left-1/2 z-10 -translate-x-1/2 animate-bounce">
        <ChevronDown className="h-7 w-7 text-white/70" />
      </div>
    </section>
  );
}
