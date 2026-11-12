"use client";

import { Suspense, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { usePathname, useSearchParams } from "next/navigation";
import { LogOut, Menu, Plus, Search, User, X } from "lucide-react";
import { toast } from "sonner";
import { useContent } from "@/lib/useContent";
import HeaderNavLink from "@/components/shared/HeaderNavLink";
import { useLogoutMutation, useMeQuery } from "@/lib/api/authApi";
import { useHydrated } from "@/hooks/use-hydrated";
import { useAppSelector } from "@/lib/hooks";
import UserAvatar from "@/components/shared/UserAvatar";
import { APP_SHELL_CLASS } from "@/lib/layout";
import { cn } from "@/lib/utils";

function navLinkActive(
  pathname: string,
  href: string,
  searchParams: URLSearchParams
): boolean {
  if (href.startsWith("/?tab=directory")) {
    return (
      (pathname === "/" && searchParams.get("tab") === "directory") ||
      pathname === "/places" ||
      pathname.startsWith("/places/")
    );
  }
  if (pathname === href) return true;
  return pathname.startsWith(`${href}/`);
}

export default function Navbar() {
  return (
    <Suspense fallback={<header className="fixed top-0 right-0 left-0 z-50 h-20 border-b border-border-subtle bg-bg/95" />}>
      <NavbarInner />
    </Suspense>
  );
}

function NavbarInner() {
  const t = useContent();
  const pathname = usePathname() ?? "";
  const searchParams = useSearchParams();
  const [mobileOpen, setMobileOpen] = useState(false);
  const hydrated = useHydrated();
  const accessToken = useAppSelector((s) => s.auth.accessToken);
  const username = useAppSelector((s) => s.auth.username);
  const isSignedIn = hydrated && Boolean(accessToken);
  const [logout, { isLoading: isLoggingOut }] = useLogoutMutation();
  const { data: me } = useMeQuery(undefined, { skip: !isSignedIn });

  const navLinks = [
    { id: "explore", label: t.nav.explore, href: "/?tab=directory" },
    { id: "assistant", label: t.nav.aiAssistant, href: "/assistant" },
    { id: "community", label: t.nav.community, href: "/qa" },
    { id: "about", label: t.nav.about, href: "/welcome" },
  ];

  const avatarUser = me ?? {
    full_name: "",
    username: username ?? "YB",
    avatar_url: "",
  };

  async function handleSignOut() {
    setMobileOpen(false);
    try {
      await logout().unwrap();
      toast.success("Signed out", { description: "See you next time." });
    } catch {
      toast.success("Signed out", { description: "Local session cleared." });
    }
  }

  return (
    <header className="fixed top-0 right-0 left-0 z-50 border-b border-border-subtle bg-bg/95 shadow-[0_4px_16px_-8px_rgba(15,23,42,0.08)] backdrop-blur-md dark:shadow-none">
      <div className={APP_SHELL_CLASS}>
        <div className="flex h-20 items-center justify-between">
          <Link href="/" className="flex shrink-0 items-center gap-2">
            <Image
              src="/images/logo.png"
              alt="Yet Bota"
              width={80}
              height={80}
              priority
              loading="eager"
              className="rounded-lg"
            />
            <span className="text-lg font-bold tracking-tight text-fg">Yet Bota</span>
          </Link>

          <nav className="hidden items-center gap-6 lg:flex xl:gap-8">
            {navLinks.map((link) => (
              <HeaderNavLink
                key={link.id}
                href={link.href}
                active={navLinkActive(pathname, link.href, searchParams)}
              >
                {link.label}
              </HeaderNavLink>
            ))}
          </nav>

          <div className="hidden items-center gap-3 lg:flex">
            {isSignedIn ? (
              <>
                <Link
                  href="/compose"
                  className="inline-flex h-9 items-center gap-1.5 rounded-full bg-brand px-3.5 text-xs font-bold text-black transition-colors hover:bg-brand-dark"
                  title="Create post"
                >
                  <Plus className="h-4 w-4" strokeWidth={2.5} />
                  <span className="hidden xl:inline">Drop</span>
                </Link>
                <Link
                  href="/search"
                  className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-border-subtle bg-overlay text-fg-muted transition-colors hover:bg-overlay-strong hover:text-fg"
                  title="Search"
                  aria-label="Search"
                >
                  <Search className="h-4 w-4" />
                </Link>
                <Link
                  href="/profile"
                  className="overflow-hidden rounded-full border border-border-subtle transition-opacity hover:opacity-90"
                  title={username ?? "Profile"}
                  aria-label="Open profile"
                >
                  <UserAvatar user={avatarUser} size="xs" />
                </Link>
                <button
                  type="button"
                  onClick={handleSignOut}
                  disabled={isLoggingOut}
                  className="text-sm text-fg-muted transition-colors hover:text-fg disabled:opacity-60"
                >
                  {t.nav.signOut}
                </button>
              </>
            ) : (
              <>
                <Link
                  href="/search"
                  className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-border-subtle bg-overlay text-fg-muted transition-colors hover:bg-overlay-strong hover:text-fg"
                  title="Search"
                  aria-label="Search"
                >
                  <Search className="h-4 w-4" />
                </Link>
                <Link href="/signin" className="text-sm text-fg-muted transition-colors hover:text-fg">
                  {t.nav.signIn}
                </Link>
                <Link
                  href="/signup"
                  className="inline-flex h-8 items-center rounded-full bg-brand px-5 text-sm font-semibold text-black hover:bg-brand-dark"
                >
                  {t.nav.joinCommunity}
                </Link>
              </>
            )}
          </div>

          <div className="flex items-center gap-2 lg:hidden">
            <Link
              href="/search"
              className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-border-subtle text-fg-muted transition-colors hover:border-fg-muted hover:text-fg"
              aria-label="Search"
            >
              <Search className="h-4 w-4" />
            </Link>
            {!isSignedIn && (
              <Link
                href="/signin"
                className="px-1 text-sm text-fg-muted transition-colors hover:text-fg"
              >
                {t.nav.signIn}
              </Link>
            )}
            <button
              type="button"
              onClick={() => setMobileOpen((v) => !v)}
              className="flex h-9 w-9 items-center justify-center rounded-lg border border-border-subtle transition-colors hover:border-fg-muted"
              aria-label="Toggle menu"
            >
              {mobileOpen ? (
                <X className="h-4 w-4 text-fg-muted" />
              ) : (
                <Menu className="h-4 w-4 text-fg-muted" />
              )}
            </button>
          </div>
        </div>
      </div>

      {mobileOpen && (
        <div className="border-t border-border-subtle bg-surface px-4 pt-3 pb-5 lg:hidden">
          <nav className="mb-3 flex flex-col gap-1">
            {navLinks.map((link) => {
              const active = navLinkActive(pathname, link.href, searchParams);
              return (
                <Link
                  key={link.id}
                  href={link.href}
                  onClick={() => setMobileOpen(false)}
                  className={cn(
                    "rounded-lg px-3 py-2.5 text-sm transition-colors",
                    active
                      ? "bg-brand/10 font-semibold text-fg"
                      : "text-fg-muted hover:bg-overlay hover:text-fg"
                  )}
                >
                  {link.label}
                </Link>
              );
            })}
          </nav>
          {isSignedIn ? (
            <div className="flex flex-col gap-1 border-t border-border-subtle pt-3">
              <Link
                href="/compose"
                onClick={() => setMobileOpen(false)}
                className="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-semibold text-brand transition-colors hover:bg-overlay"
              >
                <Plus className="h-4 w-4" />
                Drop a place
              </Link>
              <Link
                href="/search"
                onClick={() => setMobileOpen(false)}
                className="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm text-fg-muted transition-colors hover:bg-overlay hover:text-fg"
              >
                <Search className="h-4 w-4" />
                Search
              </Link>
              <Link
                href="/profile"
                onClick={() => setMobileOpen(false)}
                className="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm text-fg-muted transition-colors hover:bg-overlay hover:text-fg"
              >
                <User className="h-4 w-4" />
                {t.nav.profile}
              </Link>
              <button
                type="button"
                onClick={handleSignOut}
                disabled={isLoggingOut}
                className="flex items-center gap-2 rounded-lg px-3 py-2.5 text-left text-sm text-fg-muted transition-colors hover:bg-overlay hover:text-fg disabled:opacity-60"
              >
                <LogOut className="h-4 w-4" />
                {t.nav.signOut}
              </button>
            </div>
          ) : (
            <Link
              href="/signup"
              onClick={() => setMobileOpen(false)}
              className="inline-flex h-10 w-full items-center justify-center rounded-xl bg-brand text-sm font-semibold text-black transition-colors hover:bg-brand-dark"
            >
              {t.nav.joinCommunity}
            </Link>
          )}
        </div>
      )}
    </header>
  );
}
