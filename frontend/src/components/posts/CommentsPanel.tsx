"use client";

import { useEffect, type ReactNode } from "react";
import { X } from "lucide-react";
import CommentsView from "@/components/posts/CommentsView";
import { cn } from "@/lib/utils";

function PanelChrome({
  onClose,
  children,
  className,
  grabber,
  showClose,
}: {
  onClose?: () => void;
  children: ReactNode;
  className?: string;
  grabber?: boolean;
  showClose?: boolean;
}) {
  return (
    <div className={cn("flex min-h-0 flex-col bg-surface", className)}>
      {grabber ? (
        <div className="flex shrink-0 justify-center pt-2" aria-hidden>
          <span className="h-1 w-10 rounded-full bg-border-subtle" />
        </div>
      ) : null}
      <header className="flex shrink-0 items-center justify-between gap-3 px-4 py-2.5">
        <h2 className="text-sm font-semibold text-fg">Comments</h2>
        {showClose && onClose ? (
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg p-1.5 text-fg-muted hover:bg-surface-2 hover:text-fg"
            aria-label="Close comments"
          >
            <X className="h-4 w-4" />
          </button>
        ) : (
          <span className="w-8" aria-hidden />
        )}
      </header>
      {children}
    </div>
  );
}

/**
 * Large screens: sticky side column always shown (parent provides the grid).
 * Small screens: bottom sheet only when `mobileOpen` — post stays visible above.
 */
export default function CommentsPanel({
  postId,
  mobileOpen,
  onMobileClose,
}: {
  postId: string;
  mobileOpen: boolean;
  onMobileClose: () => void;
}) {
  useEffect(() => {
    if (!mobileOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onMobileClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [mobileOpen, onMobileClose]);

  return (
    <>
      <aside
        className="sticky top-24 hidden h-[calc(100dvh-6.5rem)] w-[22rem] min-w-[22rem] max-w-[22rem] shrink-0 flex-col overflow-hidden rounded-2xl border border-border-subtle lg:flex"
        aria-label="Comments"
      >
        <PanelChrome className="h-full min-h-0 w-full min-w-0 overflow-hidden">
          <CommentsView key={postId} postId={postId} />
        </PanelChrome>
      </aside>

      {mobileOpen ? (
        <div
          className="fixed inset-x-0 bottom-0 z-40 flex h-[min(52vh,28rem)] flex-col overflow-hidden rounded-t-2xl border border-b-0 border-border-subtle shadow-[0_-12px_40px_-20px_rgba(0,0,0,0.45)] lg:hidden"
          role="dialog"
          aria-modal="false"
          aria-label="Comments"
        >
          <PanelChrome onClose={onMobileClose} grabber showClose className="h-full">
            <CommentsView key={postId} postId={postId} />
          </PanelChrome>
        </div>
      ) : null}
    </>
  );
}
