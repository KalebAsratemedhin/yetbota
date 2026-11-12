import { redirect } from "next/navigation";

/** Legacy landing CTA — place directory lives on home with tab. */
export default function DiscoveryRedirectPage() {
  redirect("/?tab=directory");
}
