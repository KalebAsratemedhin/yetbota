"use client";

import type { ElementType } from "react";
import { Users, Bot, HandCoins, Flag } from "lucide-react";
import Reveal from "@/components/landing/Reveal";
import SectionShell from "@/components/landing/SectionShell";
import { useInView } from "@/lib/useInView";
import { useContent } from "@/lib/useContent";

const PRINCIPLE_ICONS: Record<string, ElementType> = {
  people: Users,
  ai: Bot,
  fees: HandCoins,
  digital2030: Flag,
};

export default function PrinciplesSection() {
  const t = useContent();
  const { ref, inView } = useInView<HTMLDivElement>({ threshold: 0.15 });

  return (
    <SectionShell className="bg-surface">
      <Reveal className="mb-8 text-center">
        <h2 className="mb-2 text-3xl font-extrabold tracking-tight text-fg sm:text-4xl">
          {t.principles.title}
        </h2>
        <p className="mx-auto max-w-xl text-sm text-fg-muted sm:text-base">
          {t.principles.subtitle}
        </p>
      </Reveal>

      <div
        ref={ref}
        className="mx-auto grid max-w-4xl gap-6 sm:grid-cols-2"
        style={{ perspective: "1200px" }}
      >
        {t.principles.items.map((item, i) => {
          const Icon = PRINCIPLE_ICONS[item.id] ?? Flag;
          return (
            <div
              key={item.id}
              style={{ transitionDelay: inView ? `${i * 130}ms` : "0ms" }}
              className={`origin-bottom [transform-style:preserve-3d] transition-all duration-700 ease-out ${
                inView
                  ? "translate-y-0 opacity-100 [transform:rotateX(0deg)_translateY(0)]"
                  : "translate-y-10 opacity-0 [transform:rotateX(-55deg)_translateY(2.5rem)]"
              }`}
            >
              <div className="group h-full rounded-[2rem] border border-border-subtle bg-bg p-8 transition-all duration-300 hover:-translate-y-1 hover:border-brand/35 sm:p-10">
                <div className="mb-6 flex h-14 w-14 items-center justify-center rounded-2xl bg-brand/10 transition-transform duration-300 group-hover:scale-105">
                  <Icon className="h-7 w-7 text-brand" />
                </div>
                <h3 className="mb-3 text-xl font-bold text-fg">{item.title}</h3>
                <p className="text-base leading-relaxed text-fg-faint">
                  {item.description}
                </p>
              </div>
            </div>
          );
        })}
      </div>
    </SectionShell>
  );
}
