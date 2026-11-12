"use client";

import Link from "next/link";
import { ArrowRight, Sparkles } from "lucide-react";
import Reveal from "@/components/landing/Reveal";
import SectionShell from "@/components/landing/SectionShell";
import { useContent } from "@/lib/useContent";

export default function AssistantSection() {
  const t = useContent();

  return (
    <SectionShell className="min-h-[72vh] bg-bg py-28 sm:min-h-[78vh] sm:py-36">
      <Reveal>
        <div className="overflow-hidden rounded-[2rem] border border-border-subtle bg-gradient-to-br from-brand/15 via-surface to-bg p-6 sm:p-8 lg:p-10">
          <div className="grid gap-8 lg:grid-cols-[1.1fr_0.9fr] lg:items-center">
            <div>
              <p className="mb-4 inline-flex items-center gap-2 text-[11px] font-bold tracking-[0.18em] text-brand uppercase">
                <Sparkles className="h-3.5 w-3.5" aria-hidden />
                {t.assistant.label}
              </p>
              <h2 className="mb-4 text-3xl font-extrabold tracking-tight text-fg sm:text-4xl">
                {t.assistant.title}
              </h2>
              <p className="mb-6 max-w-lg text-sm leading-relaxed text-fg-muted sm:text-base">
                Ask about neighborhoods, routes, and what’s worth knowing — then jump into
                places and posts that match.
              </p>
              <Link
                href="/assistant"
                className="inline-flex h-11 items-center gap-2 rounded-xl bg-brand px-5 text-sm font-bold text-black transition-colors hover:bg-brand-dark"
              >
                Try the assistant
                <ArrowRight className="h-4 w-4" aria-hidden />
              </Link>
            </div>

            <div className="space-y-4">
              <div className="max-w-md rounded-3xl rounded-tl-md bg-bg/85 px-5 py-4 shadow-sm">
                <p className="text-base text-fg">{t.assistant.exampleQuery}</p>
              </div>
              <div className="rounded-3xl rounded-tl-md border border-brand/25 bg-brand/10 px-5 py-5">
                <p className="mb-3 text-[10px] font-bold tracking-widest text-brand uppercase">
                  {t.assistant.label}
                </p>
                <p
                  className="text-base leading-relaxed text-fg-muted"
                  dangerouslySetInnerHTML={{
                    __html: t.assistant.exampleResponse.replace(
                      /\*\*(.*?)\*\*/g,
                      '<strong class="text-fg">$1</strong>'
                    ),
                  }}
                />
              </div>
            </div>
          </div>
        </div>
      </Reveal>
    </SectionShell>
  );
}
