import { InputHTMLAttributes } from "react";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string | null;
}

export default function Input({
  label,
  error,
  id,
  className = "",
  ...props
}: InputProps) {
  return (
    <div className="space-y-1.5">
      <label
        htmlFor={id}
        className="block text-sm font-semibold text-bark-light"
      >
        {label}
      </label>
      <input
        id={id}
        className={`
          w-full min-h-[48px] px-4 py-3
          rounded-xl border border-sand bg-white
          text-bark text-base placeholder:text-bark-light/50
          transition-all duration-200
          focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest
          disabled:bg-cream-dark disabled:cursor-not-allowed
          ${error ? "border-terracotta ring-1 ring-terracotta/30" : ""}
          ${className}
        `}
        {...props}
      />
      {error && (
        <p className="text-sm text-terracotta font-medium">{error}</p>
      )}
    </div>
  );
}
