"use client";

import { Navbar } from "@/components/landing";
import ChatPanel from "@/components/assistant/ChatPanel";
import RequireAuth from "@/components/auth/RequireAuth";

export default function AssistantPage() {
  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg text-fg">
        <Navbar />
        <div className="pt-20">
          <ChatPanel />
        </div>
      </main>
    </RequireAuth>
  );
}
