"use client";

import { useCallback } from "react";
import { usePathname, useRouter } from "next/navigation";
import { toast } from "sonner";
import { useAppSelector } from "@/lib/hooks";

/** Returns false and sends the user to sign-in when logged out. */
export function useRequireSignIn() {
  const router = useRouter();
  const pathname = usePathname();
  const accessToken = useAppSelector((s) => s.auth.accessToken);
  const isSignedIn = Boolean(accessToken);

  const requireSignIn = useCallback(
    (message = "Sign in to continue") => {
      if (accessToken) return true;
      const next = pathname && pathname.startsWith("/") ? pathname : "/";
      const href = `/signin?next=${encodeURIComponent(next)}`;
      toast.message(message, {
        action: {
          label: "Sign in",
          onClick: () => router.push(href),
        },
      });
      router.push(href);
      return false;
    },
    [accessToken, pathname, router]
  );

  return { isSignedIn, requireSignIn };
}
