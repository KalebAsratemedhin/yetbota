"use client";

import Link from "next/link";
import Image from "next/image";
import { ChevronLeft } from "lucide-react";

interface AuthCardProps {
  title: string;
  children: React.ReactNode;
  backHref?: string;
}

export default function AuthCard({ title, children, backHref = "/" }: AuthCardProps) {
  return (
    <div className="w-full max-w-sm">
      <div className="mt-3 mb-8 flex items-center justify-between">
        <Link
          href={backHref}
          className="flex items-center gap-1.5 text-sm text-fg-muted transition-colors hover:text-fg"
        >
          <ChevronLeft className="h-4 w-4" />
          <span>Back</span>
        </Link>
        <span className="text-sm font-semibold text-fg">{title}</span>
        <div className="w-14" />
      </div>

      <div className="mb-8 flex flex-col items-center">
        <div className="mb-1 flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl shadow-[0_0_40px_#22C55E33]">
          <Image
            src="/images/logo.png"
            alt="Yet Bota"
            width={80}
            height={80}
            className="rounded-sm"
          />
        </div>
      </div>

      {children}
    </div>
  );
}
