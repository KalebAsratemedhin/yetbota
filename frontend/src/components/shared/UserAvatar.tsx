import { cn } from "@/lib/utils";
import type { UserInfo } from "@/lib/auth/types";

function initialsFrom(user: Pick<UserInfo, "full_name" | "username">): string {
  const fromName = user.full_name
    .split(" ")
    .filter(Boolean)
    .map((n) => n[0])
    .join("");
  return (fromName || user.username || "YB").slice(0, 2).toUpperCase();
}

const sizeClass = {
  xs: "h-9 w-9 text-[11px] rounded-full",
  sm: "h-9 w-9 text-xs rounded-xl",
  md: "h-12 w-12 text-sm rounded-2xl",
  lg: "h-24 w-24 text-xl rounded-3xl sm:h-32 sm:w-32 sm:text-2xl lg:h-40 lg:w-40 lg:text-3xl",
} as const;

export default function UserAvatar({
  user,
  size = "md",
  className,
}: {
  user: Pick<UserInfo, "full_name" | "username" | "avatar_url">;
  size?: keyof typeof sizeClass;
  className?: string;
}) {
  const url = user.avatar_url?.trim();
  return (
    <div
      className={cn(
        "relative flex shrink-0 items-center justify-center overflow-hidden bg-[#e8b89a]",
        sizeClass[size],
        className
      )}
    >
      {url ? (
        // eslint-disable-next-line @next/next/no-img-element -- Cloudinary URLs; avoid remotePatterns churn
        <img src={url} alt="" className="h-full w-full object-cover" />
      ) : (
        <span className="font-bold text-fg/60">{initialsFrom(user)}</span>
      )}
    </div>
  );
}
