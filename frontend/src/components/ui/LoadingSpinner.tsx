import { Loader2 } from "lucide-react";

interface LoadingSpinnerProps {
  message?: string;
  variant?: "page" | "inline";
}

export default function LoadingSpinner({
  message,
  variant = "page",
}: LoadingSpinnerProps) {
  if (variant === "inline") {
    return (
      <span className="inline-flex items-center gap-2 text-bark-light">
        <Loader2 className="h-4 w-4 animate-spin" />
        {message && <span className="text-sm">{message}</span>}
      </span>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-[200px] gap-3">
      <Loader2 className="h-8 w-8 animate-spin text-forest" />
      {message && (
        <p className="text-bark-light text-base font-medium">{message}</p>
      )}
    </div>
  );
}
