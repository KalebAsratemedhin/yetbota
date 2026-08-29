"use client";

import Link from "next/link";
import { Send, Pencil, HelpCircle, Sparkles } from "lucide-react";
import { useContent } from "@/lib/useContent";
import Reveal from "@/components/landing/Reveal";

export default function AssistantSection() {
  const t = useContent();

  return (
    <section className="bg-bg px-6 py-14">
      <div className="mx-auto grid max-w-7xl gap-6 md:grid-cols-2">
        <Reveal direction="left" className="h-full">
          <div className="h-full rounded-3xl border border-border-subtle bg-bg p-6 transition-all duration-300 hover:border-brand/40 hover:shadow-xl hover:shadow-brand/10">
            <div className="mb-6 flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-brand/20">
                <Sparkles className="h-5 w-5 text-brand" />
              </div>
              <h3 className="text-lg font-semibold text-fg">{t.assistant.title}</h3>
            </div>

            <div className="mb-6 space-y-4">
              <div className="max-w-xs rounded-2xl rounded-tl-sm bg-surface px-4 py-3">
                <p className="text-sm text-fg">{t.assistant.exampleQuery}</p>
              </div>

              <div className="rounded-2xl rounded-tl-sm border border-brand/20 bg-brand/10 px-4 py-3">
                <p className="mb-2 text-[10px] font-bold tracking-widest text-brand uppercase">
                  {t.assistant.label}
                </p>
                <p
                  className="text-sm leading-relaxed text-fg-muted"
                  dangerouslySetInnerHTML={{
                    __html: t.assistant.exampleResponse.replace(
                      /\*\*(.*?)\*\*/g,
                      '<strong class="text-fg">$1</strong>'
                    ),
                  }}
                />
              </div>
            </div>

            <div className="flex items-center gap-3 rounded-2xl border border-border-subtle bg-surface-2 px-4 py-3">
              <input
                type="text"
                placeholder={t.assistant.placeholder}
                className="flex-1 bg-transparent text-sm text-fg-muted outline-none placeholder:text-gray-600"
                readOnly
              />
              <Link
                href="/assistant"
                aria-label={t.nav.aiAssistant}
                className="flex h-8 w-8 items-center justify-center rounded-full bg-brand transition-colors hover:bg-brand-dark"
              >
                <Send className="h-4 w-4 text-black" />
              </Link>
            </div>
          </div>
        </Reveal>

        <Reveal direction="right" delay={150} className="h-full">
          <div className="flex h-full flex-col rounded-3xl border border-border-subtle bg-bg p-6 transition-all duration-300 hover:border-brand/40 hover:shadow-xl hover:shadow-brand/10">
            <div className="mb-6 flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-brand/20">
                <HelpCircle className="h-5 w-5 text-brand" />
              </div>
              <h3 className="text-lg font-semibold text-fg">{t.community.helpTitle}</h3>
            </div>

            <div className="mb-6 flex-1 rounded-2xl border border-border-subtle bg-surface-2 p-5">
              <div className="mb-4 flex items-center gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gray-700 text-sm font-bold text-white">
                  A
                </div>
                <div>
                  <p className="text-sm font-semibold text-fg">{t.community.user}</p>
                  <p className="text-[10px] font-bold tracking-widest text-brand uppercase">
                    {t.community.trendingQuestion}
                  </p>
                </div>
              </div>
              <p className="mb-4 text-base leading-relaxed text-fg">{t.community.exampleQuestion}</p>
              <div className="flex items-center justify-between">
                <div className="flex -space-x-2">
                  {["B", "S"].map((initial) => (
                    <div
                      key={initial}
                      className="flex h-7 w-7 items-center justify-center rounded-full border-2 border-surface bg-gray-600 text-[10px] font-bold text-fg"
                    >
                      {initial}
                    </div>
                  ))}
                </div>
                <Link
                  href="/ask"
                  className="inline-flex items-center rounded-full bg-brand px-5 py-1.5 text-sm font-semibold text-black transition-colors hover:bg-brand-dark"
                >
                  <Pencil className="mr-1.5 h-3 w-3" />
                  {t.community.helpOut}
                </Link>
              </div>
            </div>

            <Link
              href="/qa"
              className="text-center text-sm text-fg-muted transition-colors hover:text-fg"
            >
              {t.community.browseQuestions} →
            </Link>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
