"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useHydrated } from "@/hooks/use-hydrated";
import { useAppSelector } from "@/lib/hooks";

export default function RequireAuth({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const accessToken = useAppSelector((s) => s.auth.accessToken);
  const hydrated = useHydrated();

  useEffect(() => {
    if (!hydrated) return;
    if (accessToken) return;
    const next = pathname && pathname.startsWith("/") ? pathname : "/profile";
    router.replace(`/signin?next=${encodeURIComponent(next)}`);
  }, [hydrated, accessToken, pathname, router]);

  if (!hydrated || !accessToken) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center bg-bg text-sm text-fg-faint">
        {hydrated && !accessToken ? "Redirecting to sign in…" : "Loading…"}
      </div>
    );
  }

  return <>{children}</>;
}
