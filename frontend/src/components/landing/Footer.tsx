"use client";

import Link from "next/link";
import Image from "next/image";
import { useContent } from "@/lib/useContent";

export default function Footer() {
  const t = useContent();
  const navGroups = [t.footer.nav.discovery, t.footer.nav.community];

  return (
    <footer className="border-t border-border-subtle bg-bg px-6 pt-20 pb-10 sm:pt-24">
      <div className="mx-auto max-w-6xl">
        <div className="mb-16 grid grid-cols-2 gap-12 md:grid-cols-3">
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
            <p className="mb-5 max-w-55 text-sm leading-relaxed text-fg-faint">
              {t.footer.tagline}
            </p>
          </div>

          {navGroups.map((group) => (
            <div key={group.title}>
              <h4 className="mb-5 text-sm font-semibold text-fg">{group.title}</h4>
              <ul className="space-y-3">
                {group.links.map((link) => (
                  <li key={link.href}>
                    <Link
                      href={link.href}
                      className="text-sm text-fg-faint transition-colors hover:text-fg"
                    >
                      {link.label}
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
