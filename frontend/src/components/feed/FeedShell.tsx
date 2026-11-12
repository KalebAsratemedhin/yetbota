"use client";

import type { ReactNode } from "react";
import TrendingRail from "@/components/feed/TrendingRail";
import { APP_SHELL_CLASS } from "@/lib/layout";
import { cn } from "@/lib/utils";

/**
 * Home / post detail shell: shared padding, top spacing, trending rail + main column.
 */
export default function FeedShell({
  children,
  className,
  railTop,
}: {
  children: ReactNode;
  className?: string;
  /** Rendered at the top of the trending sidebar (home explore tabs). */
  railTop?: ReactNode;
}) {
  return (
    <div
      className={cn(
        APP_SHELL_CLASS,
        "flex items-start gap-6 pt-4 sm:gap-8 sm:pt-5 lg:gap-10 lg:pt-6",
        className
      )}
    >
      <div className="sticky top-24 hidden w-60 shrink-0 self-start xl:block">
        <TrendingRail top={railTop} />
      </div>
      <div className="min-w-0 flex-1">{children}</div>
    </div>
  );
}
