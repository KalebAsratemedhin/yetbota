"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { ExternalLink, Send, Sparkles } from "lucide-react";
import { toast } from "sonner";
import { useChatMutation } from "@/lib/api/aiApi";
import { chatSourceHref, chatSourceLabel } from "@/lib/ai/helpers";
import type { ChatSource, ChatTurn } from "@/lib/ai/types";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { cn } from "@/lib/utils";

type UiMessage = {
  id: string;
  role: "user" | "assistant";
  content: string;
  sources?: ChatSource[];
  model?: string;
};

const MAX_HISTORY = 20;
const SUGGESTIONS = [
  "Best coffee near Piazza?",
  "Quiet spots to work in Addis?",
  "Where can I park near Friendship Park?",
];

function newId() {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

function getChatErrorMessage(error: unknown): string {
  if (error && typeof error === "object" && "data" in error) {
    const data = (error as { data: unknown }).data;
    if (data && typeof data === "object" && "detail" in data) {
      const detail = (data as { detail: unknown }).detail;
      if (typeof detail === "string" && detail.trim()) return detail;
    }
  }
  return getApiErrorMessage(
    error,
    "Could not reach the assistant. Is ai-service running?"
  );
}

function formatReply(text: string): string {
  return text.replace(/\*\*(.*?)\*\*/g, '<strong class="text-fg">$1</strong>');
}

function SourcesList({ sources }: { sources: ChatSource[] }) {
  if (sources.length === 0) return null;
  return (
    <ul className="mt-3 space-y-2 border-t border-brand/15 pt-3">
      <li className="text-[10px] font-bold tracking-widest text-brand uppercase">
        Sources
      </li>
      {sources.map((s) => {
        const href = chatSourceHref(s);
        const label = chatSourceLabel(s);
        const snippet = s.text?.trim();
        const inner = (
          <>
            <span className="font-semibold text-fg">{label}</span>
            <span className="ml-1.5 text-[10px] font-medium tracking-wide text-fg-faint uppercase">
              {s.kind}
            </span>
            {snippet ? (
              <p className="mt-0.5 line-clamp-2 text-xs text-fg-muted">{snippet}</p>
            ) : null}
          </>
        );
        return (
          <li key={`${s.kind}-${s.entity_id}`}>
            {href ? (
              <Link
                href={href}
                className="block rounded-xl border border-border-subtle bg-bg/60 px-3 py-2 transition-colors hover:border-brand/40 hover:bg-bg"
              >
                <span className="inline-flex items-start gap-1.5">
                  <span className="min-w-0 flex-1">{inner}</span>
                  <ExternalLink className="mt-0.5 h-3.5 w-3.5 shrink-0 text-fg-faint" />
                </span>
              </Link>
            ) : (
              <div className="rounded-xl border border-border-subtle bg-bg/60 px-3 py-2">
                {inner}
              </div>
            )}
          </li>
        );
      })}
    </ul>
  );
}

export default function ChatPanel() {
  const [messages, setMessages] = useState<UiMessage[]>([]);
  const [input, setInput] = useState("");
  const [chat, { isLoading }] = useChatMutation();
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [messages, isLoading]);

  async function send(raw: string) {
    const message = raw.trim().slice(0, 4000);
    if (!message || isLoading) return;

    const userMsg: UiMessage = { id: newId(), role: "user", content: message };
    const prior = messages;
    setMessages((prev) => [...prev, userMsg]);
    setInput("");

    const history: ChatTurn[] = prior
      .slice(-MAX_HISTORY)
      .map(({ role, content }) => ({ role, content }));

    try {
      const res = await chat({ message, history }).unwrap();
      setMessages((prev) => [
        ...prev,
        {
          id: newId(),
          role: "assistant",
          content: res.reply,
          sources: res.sources,
          model: res.model,
        },
      ]);
    } catch (err) {
      setMessages((prev) => prev.filter((m) => m.id !== userMsg.id));
      setInput(message);
      toast.error("Assistant unavailable", {
        description: getChatErrorMessage(err),
      });
    } finally {
      inputRef.current?.focus();
    }
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    void send(input);
  }

  function onKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      void send(input);
    }
  }

  const empty = messages.length === 0;

  return (
    <div className="mx-auto flex h-[calc(100dvh-5rem)] w-full max-w-3xl flex-col px-4 sm:px-6">
      <header className="shrink-0 border-b border-border-subtle py-5">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-brand/20">
            <Sparkles className="h-5 w-5 text-brand" />
          </div>
          <div>
            <h1 className="text-lg font-bold tracking-tight text-fg">
              Ask Yet Bota
            </h1>
            <p className="text-sm text-fg-muted">
              Answers from community posts and Q&A around Addis.
            </p>
          </div>
        </div>
      </header>

      <div className="yb-scrollbar min-h-0 flex-1 space-y-4 overflow-y-auto py-6">
        {empty ? (
          <div className="flex h-full flex-col items-center justify-center gap-6 px-2 text-center">
            <p className="max-w-sm text-sm text-fg-muted">
              Ask about places, parking, coffee, or anything locals have shared.
            </p>
            <ul className="flex w-full max-w-md flex-col gap-2">
              {SUGGESTIONS.map((q) => (
                <li key={q}>
                  <button
                    type="button"
                    onClick={() => void send(q)}
                    className="w-full rounded-2xl border border-border-subtle bg-surface px-4 py-3 text-left text-sm text-fg transition-colors hover:border-brand/40 hover:bg-surface-2"
                  >
                    {q}
                  </button>
                </li>
              ))}
            </ul>
          </div>
        ) : (
          messages.map((m) =>
            m.role === "user" ? (
              <div key={m.id} className="flex justify-end">
                <div className="max-w-[85%] rounded-2xl rounded-br-sm bg-surface px-4 py-3">
                  <p className="whitespace-pre-wrap text-sm text-fg">{m.content}</p>
                </div>
              </div>
            ) : (
              <div key={m.id} className="flex justify-start">
                <div className="max-w-[90%] rounded-2xl rounded-tl-sm border border-brand/20 bg-brand/10 px-4 py-3">
                  <p className="mb-2 text-[10px] font-bold tracking-widest text-brand uppercase">
                    Yet Bota Assistant
                  </p>
                  <p
                    className="text-sm leading-relaxed whitespace-pre-wrap text-fg-muted"
                    dangerouslySetInnerHTML={{
                      __html: formatReply(m.content),
                    }}
                  />
                  {m.sources ? <SourcesList sources={m.sources} /> : null}
                </div>
              </div>
            )
          )
        )}

        {isLoading ? (
          <div className="flex justify-start">
            <div className="rounded-2xl rounded-tl-sm border border-brand/20 bg-brand/10 px-4 py-3">
              <p className="text-[10px] font-bold tracking-widest text-brand uppercase">
                Yet Bota Assistant
              </p>
              <p className="mt-2 text-sm text-fg-muted">Searching places…</p>
            </div>
          </div>
        ) : null}
        <div ref={bottomRef} />
      </div>

      <form
        onSubmit={onSubmit}
        className="shrink-0 border-t border-border-subtle bg-bg py-4 pb-[max(1rem,env(safe-area-inset-bottom))]"
      >
        <div className="flex items-end gap-3 rounded-2xl border border-border-subtle bg-surface-2 px-3 py-2.5 focus-within:border-brand/40">
          <textarea
            ref={inputRef}
            rows={1}
            value={input}
            onChange={(e) => setInput(e.target.value.slice(0, 4000))}
            onKeyDown={onKeyDown}
            placeholder="Ask about a place…"
            disabled={isLoading}
            className={cn(
              "yb-scrollbar max-h-32 min-h-[2.25rem] flex-1 resize-none bg-transparent py-1.5 text-sm text-fg outline-none",
              "placeholder:text-fg-faint disabled:opacity-60"
            )}
          />
          <button
            type="submit"
            disabled={isLoading || !input.trim()}
            aria-label="Send"
            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-brand text-black transition-colors hover:bg-brand-dark disabled:opacity-40"
          >
            <Send className="h-4 w-4" />
          </button>
        </div>
        <p className="mt-2 text-center text-[11px] text-fg-faint">
          Enter to send · Shift+Enter for a new line
        </p>
      </form>
    </div>
  );
}
