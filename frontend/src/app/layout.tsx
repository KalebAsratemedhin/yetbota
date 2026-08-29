import type { Metadata } from "next";
import { Geist_Mono, Plus_Jakarta_Sans } from "next/font/google";
import { StoreProvider } from "@/components/providers/store-provider";
import "./globals.css";

const jakarta = Plus_Jakarta_Sans({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-sans",
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Yet Bota – Digital Address & Local Discovery",
  description:
    "Community-contributed location discovery platform aligned with Ethiopia's Digital 2030 strategy.",
};

const themeInitScript = `(() => {
  try {
    const saved = localStorage.getItem('yetbota.theme');
    const isDark = saved ? saved === 'dark' : true;
    document.documentElement.classList.toggle('dark', isDark);
  } catch (_) {
    document.documentElement.classList.add('dark');
  }
})();`;

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={`${jakarta.variable} ${geistMono.variable} h-full antialiased`}
    >
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body className={`${jakarta.className} flex min-h-full flex-col bg-bg text-fg`}>
        <StoreProvider>{children}</StoreProvider>
      </body>
    </html>
  );
}
