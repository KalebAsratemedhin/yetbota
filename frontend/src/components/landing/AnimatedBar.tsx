"use client";

import { useInView } from "@/lib/useInView";

export default function AnimatedBar({
  value,
  className = "",
}: {
  value: number;
  className?: string;
}) {
  const { ref, inView } = useInView<HTMLDivElement>();

  return (
    <div ref={ref} className={`h-1.5 w-full overflow-hidden rounded-full bg-surface ${className}`}>
      <div
        className="h-1.5 rounded-full bg-brand transition-[width] duration-[1400ms] ease-out"
        style={{ width: inView ? `${value}%` : "0%" }}
      />
    </div>
  );
}
