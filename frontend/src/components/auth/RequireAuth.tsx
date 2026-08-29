"use client";

import { useEffect, useSyncExternalStore } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAppSelector } from "@/lib/hooks";

const emptySubscribe = () => () => {};

export default function RequireAuth({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const accessToken = useAppSelector((s) => s.auth.accessToken);
  /** false on server + first SSR HTML; true after client hydrate */
  const isClient = useSyncExternalStore(emptySubscribe, () => true, () => false);

  useEffect(() => {
    if (!isClient) return;
    if (accessToken) return;
    const next = pathname && pathname.startsWith("/") ? pathname : "/profile";
    router.replace(`/signin?next=${encodeURIComponent(next)}`);
  }, [isClient, accessToken, pathname, router]);

  if (!isClient || !accessToken) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center bg-bg text-sm text-fg-faint">
        {isClient && !accessToken ? "Redirecting to sign in…" : "Loading…"}
      </div>
    );
  }

  return <>{children}</>;
}
