import { redirect } from "next/navigation";

/** Legacy landing CTA — place directory lives at /places. */
export default function DiscoveryRedirectPage() {
  redirect("/places");
}
