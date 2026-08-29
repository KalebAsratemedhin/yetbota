import { cn } from "@/lib/utils";

export function ProfileSection({
  title,
  description,
  children,
  className,
  action,
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
  action?: React.ReactNode;
}) {
  return (
    <section
      className={cn(
        "rounded-2xl border border-border-subtle/80 bg-surface/80 p-5 backdrop-blur-sm sm:p-6",
        className
      )}
    >
      <div className="mb-5 flex items-start justify-between gap-4">
        <div className="min-w-0">
          <h2 className="text-base font-semibold tracking-tight text-fg sm:text-lg">{title}</h2>
          {description && (
            <p className="mt-1 text-sm leading-relaxed text-fg-faint">{description}</p>
          )}
        </div>
        {action}
      </div>
      {children}
    </section>
  );
}

export function ComingSoonChip() {
  return (
    <span className="shrink-0 rounded-full bg-overlay px-2.5 py-1 text-[10px] font-semibold tracking-wide text-fg-faint uppercase">
      Soon
    </span>
  );
}
