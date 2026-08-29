"use client";

import { useInView } from "@/lib/useInView";

interface RevealProps {
  children: React.ReactNode;
  className?: string;
  delay?: number;
  direction?: "up" | "left" | "right" | "scale";
}

const HIDDEN: Record<NonNullable<RevealProps["direction"]>, string> = {
  up: "opacity-0 translate-y-10",
  left: "opacity-0 -translate-x-10",
  right: "opacity-0 translate-x-10",
  scale: "opacity-0 scale-90",
};

export default function Reveal({
  children,
  className = "",
  delay = 0,
  direction = "up",
}: RevealProps) {
  const { ref, inView } = useInView<HTMLDivElement>();

  return (
    <div
      ref={ref}
      style={{ transitionDelay: inView ? `${delay}ms` : "0ms" }}
      className={`transition-all duration-700 ease-out will-change-transform ${
        inView ? "opacity-100 translate-x-0 translate-y-0 scale-100" : HIDDEN[direction]
      } ${className}`}
    >
      {children}
    </div>
  );
}
