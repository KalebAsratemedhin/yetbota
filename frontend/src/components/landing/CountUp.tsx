"use client";

import { useEffect, useState } from "react";
import { useInView } from "@/lib/useInView";

export default function CountUp({
  value,
  className = "",
  duration = 1600,
}: {
  value: string;
  className?: string;
  duration?: number;
}) {
  const { ref, inView } = useInView<HTMLSpanElement>();

  const m = value.match(/^(\D*)([\d.]+)(.*)$/);
  const hasNumber = m !== null;
  const prefix = m ? m[1] : "";
  const target = m ? parseFloat(m[2]) : 0;
  const suffix = m ? m[3] : value;
  const decimals = m && m[2].includes(".") ? m[2].split(".")[1].length : 0;

  const [display, setDisplay] = useState(0);

  useEffect(() => {
    if (!inView || !hasNumber) return;
    let raf = 0;
    const start = performance.now();
    const tick = (now: number) => {
      const p = Math.min((now - start) / duration, 1);
      const eased = 1 - Math.pow(1 - p, 3);
      setDisplay(target * eased);
      if (p < 1) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [inView, hasNumber, target, duration]);

  return (
    <span ref={ref} className={className}>
      {hasNumber ? `${prefix}${(inView ? display : 0).toFixed(decimals)}${suffix}` : value}
    </span>
  );
}
