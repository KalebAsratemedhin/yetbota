import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export default function SectionShell({
  children,
  className,
  innerClassName,
}: {
  children: ReactNode;
  className?: string;
  innerClassName?: string;
}) {
  return (
    <section
      className={cn(
        "flex min-h-[48vh] flex-col justify-center px-4 py-16 sm:min-h-[52vh] sm:px-6 sm:py-20 lg:px-8",
        className
      )}
    >
      <div className={cn("mx-auto w-full max-w-6xl", innerClassName)}>{children}</div>
    </section>
  );
}
