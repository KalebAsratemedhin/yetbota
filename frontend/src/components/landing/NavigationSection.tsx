"use client";

import Image from "next/image";
import { CheckCircle2 } from "lucide-react";
import { useContent } from "@/lib/useContent";
import Reveal from "@/components/landing/Reveal";

export default function NavigationSection() {
  const t = useContent();

  return (
    <section className="bg-brand/5 px-6 py-16 dark:bg-surface">
      <div className="mx-auto grid max-w-7xl items-center gap-12 md:grid-cols-2">
        <Reveal direction="left">
          <p className="mb-4 text-xs font-bold tracking-widest text-brand uppercase">
            {t.navigation.badge}
          </p>
          <h2 className="mb-6 text-4xl leading-tight font-extrabold text-fg md:text-5xl">
            {t.navigation.title}
          </h2>
          <p className="mb-8 text-base leading-relaxed text-fg-muted">{t.navigation.description}</p>
          <ul className="space-y-3">
            {t.navigation.features.map((feature) => (
              <li key={feature} className="flex items-center gap-3">
                <CheckCircle2 className="h-5 w-5 shrink-0 text-brand" />
                <span className="text-sm text-fg-muted">{feature}</span>
              </li>
            ))}
          </ul>
        </Reveal>

        <Reveal direction="right" delay={150} className="relative">
          <div className="group relative h-80 w-full overflow-hidden rounded-3xl border border-border-subtle md:h-96">
            <Image
              src="/images/yetbota-bg-1.jpeg"
              alt="Local navigation"
              fill
              className="object-cover transition-transform duration-500 group-hover:scale-105"
              sizes="(max-width: 768px) 100vw, 50vw"
            />
          </div>
        </Reveal>
      </div>
    </section>
  );
}
