"use client";

import { Navbar } from "@/components/landing";
import ChatPanel from "@/components/assistant/ChatPanel";

export default function AssistantPage() {
  return (
    <main className="min-h-screen overflow-x-clip bg-bg text-fg">
      <Navbar />
      <div className="pt-20">
        <ChatPanel />
      </div>
    </main>
  );
}
