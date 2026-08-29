"use client";

import type { ElementType } from "react";
import { Users, Bot, HandCoins, Flag } from "lucide-react";
import { useContent } from "@/lib/useContent";
import Reveal from "@/components/landing/Reveal";
import { useInView } from "@/lib/useInView";

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
    <section className="bg-bg px-6 py-16">
      <div className="mx-auto max-w-7xl">
        <Reveal className="mb-14 text-center">
          <h2 className="mb-3 text-4xl font-extrabold text-fg">{t.principles.title}</h2>
          <p className="text-base text-fg-faint">{t.principles.subtitle}</p>
        </Reveal>

        <div
          ref={ref}
          className="mx-auto grid max-w-3xl gap-5 sm:grid-cols-2"
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
                <div className="group h-full rounded-3xl border border-border-subtle bg-bg p-6 transition-all duration-300 hover:-translate-y-1.5 hover:border-brand/40 hover:shadow-xl hover:shadow-brand/10">
                  <div className="mb-5 flex h-12 w-12 items-center justify-center rounded-2xl bg-brand/10 transition-all duration-300 group-hover:scale-110 group-hover:rotate-6 group-hover:bg-brand/20">
                    <Icon className="h-6 w-6 text-brand transition-transform duration-300 group-hover:scale-110" />
                  </div>
                  <h3 className="mb-3 text-base font-bold text-fg">{item.title}</h3>
                  <p className="text-sm leading-relaxed text-fg-faint">{item.description}</p>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
