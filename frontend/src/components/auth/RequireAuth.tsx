"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAppSelector } from "@/lib/hooks";

export default function RequireAuth({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const accessToken = useAppSelector((s) => s.auth.accessToken);

  useEffect(() => {
    if (accessToken) return;
    const next = pathname && pathname.startsWith("/") ? pathname : "/profile";
    router.replace(`/signin?next=${encodeURIComponent(next)}`);
  }, [accessToken, pathname, router]);

  if (!accessToken) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center bg-bg text-sm text-fg-faint">
        Redirecting to sign in…
      </div>
    );
  }

  return <>{children}</>;
}
