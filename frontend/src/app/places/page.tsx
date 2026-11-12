import { redirect } from "next/navigation";

/** Explore Place Directory lives on `/` with `?tab=directory`. */
export default function PlacesPage() {
  redirect("/?tab=directory");
}
