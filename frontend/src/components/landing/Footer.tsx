"use client";

import Link from "next/link";
import Image from "next/image";
import { Globe, Share2 } from "lucide-react";
import { useContent } from "@/lib/useContent";

export default function Footer() {
  const t = useContent();
  const navGroups = [t.footer.nav.discovery, t.footer.nav.community];

  return (
    <footer className="border-t border-border-subtle bg-bg px-6 pt-16 pb-8">
      <div className="mx-auto max-w-7xl">
        <div className="mb-14 grid grid-cols-2 gap-10 md:grid-cols-3">
          <div className="col-span-2 md:col-span-1">
            <div className="mb-4 flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center overflow-hidden rounded-lg">
                <Image
                  src="/images/logo.png"
                  alt="Yet Bota"
                  width={22}
                  height={22}
                  className="rounded-sm"
                />
              </div>
              <span className="text-lg font-bold text-fg">Yet Bota</span>
            </div>
            <p className="mb-5 max-w-55 text-sm leading-relaxed text-fg-faint">{t.footer.tagline}</p>
            <div className="flex items-center gap-3">
              <button
                type="button"
                className="flex h-9 w-9 items-center justify-center rounded-full border border-border-subtle"
                aria-label="Language"
              >
                <Globe className="h-4 w-4 text-fg-faint" />
              </button>
              <button
                type="button"
                className="flex h-9 w-9 items-center justify-center rounded-full border border-border-subtle"
                aria-label="Share"
              >
                <Share2 className="h-4 w-4 text-fg-faint" />
              </button>
            </div>
          </div>

          {navGroups.map((group) => (
            <div key={group.title}>
              <h4 className="mb-5 text-sm font-semibold text-fg">{group.title}</h4>
              <ul className="space-y-3">
                {group.links.map((link) => (
                  <li key={link}>
                    <Link
                      href="#"
                      className="relative inline-flex items-center text-sm text-fg-faint transition-colors hover:text-green-700 after:absolute after:left-1/2 after:-bottom-1.5 after:h-0.5 after:w-0 after:bg-green-700 after:transition-all after:duration-300 after:content-[''] hover:after:w-full after:-translate-x-1/2 dark:hover:text-green-500 dark:after:bg-green-500"
                    >
                      {link}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="border-t border-border-subtle pt-6">
          <p className="text-center text-xs tracking-widest text-fg-faint uppercase md:text-left">
            {t.footer.copyright}
          </p>
        </div>
      </div>
    </footer>
  );
}
