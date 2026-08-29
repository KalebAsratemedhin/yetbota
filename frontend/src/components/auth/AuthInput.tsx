"use client";

import { forwardRef, useId, useState } from "react";
import { Eye, EyeOff } from "lucide-react";
import { cn } from "@/lib/utils";

interface AuthInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  rightLabel?: React.ReactNode;
  error?: string;
}

const AuthInput = forwardRef<HTMLInputElement, AuthInputProps>(
  ({ label, rightLabel, error, className, type, id, name, autoComplete, ...props }, ref) => {
    const [showPassword, setShowPassword] = useState(false);
    const isPassword = type === "password";
    const reactId = useId();
    const inputId = id ?? `auth-input-${reactId}`;
    const inputName = name ?? autoComplete ?? undefined;

    return (
      <div className="w-full">
        {(label || rightLabel) && (
          <div className="mb-1.5 flex items-center justify-between">
            {label && (
              <label htmlFor={inputId} className="text-sm font-medium text-fg-muted">
                {label}
              </label>
            )}
            {rightLabel && <div className="text-sm">{rightLabel}</div>}
          </div>
        )}
        <div className="relative">
          <input
            ref={ref}
            id={inputId}
            name={inputName}
            autoComplete={autoComplete}
            type={isPassword && showPassword ? "text" : type}
            className={cn(
              "w-full rounded-xl border border-border-subtle bg-surface-2 px-4 py-3 text-sm text-fg outline-none placeholder:text-gray-600",
              "transition-all focus:border-brand/50 focus:ring-1 focus:ring-brand/20",
              isPassword && "pr-11",
              error && "border-red-500/50",
              className
            )}
            {...props}
          />
          {isPassword && (
            <button
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              className="absolute top-1/2 right-3 -translate-y-1/2 text-fg-faint transition-colors hover:text-fg-muted"
              aria-label={showPassword ? "Hide password" : "Show password"}
            >
              {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </button>
          )}
        </div>
        {error && <p className="mt-1 text-xs text-red-400">{error}</p>}
      </div>
    );
  }
);

AuthInput.displayName = "AuthInput";
export default AuthInput;
