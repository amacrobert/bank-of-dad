import { SelectHTMLAttributes } from "react";

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string;
  error?: string | null;
}

export default function Select({
  label,
  error,
  id,
  className = "",
  children,
  ...props
}: SelectProps) {
  return (
    <div className="space-y-1.5">
      <label
        htmlFor={id}
        className="block text-sm font-semibold text-bark-light"
      >
        {label}
      </label>
      <select
        id={id}
        className={`
          w-full min-h-[48px] px-4 py-3
          rounded-xl border border-sand bg-white
          text-bark text-base
          transition-all duration-200
          focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest
          disabled:bg-cream-dark disabled:cursor-not-allowed
          ${error ? "border-terracotta ring-1 ring-terracotta/30" : ""}
          ${className}
        `}
        {...props}
      >
        {children}
      </select>
      {error && (
        <p className="text-sm text-terracotta font-medium">{error}</p>
      )}
    </div>
  );
}
