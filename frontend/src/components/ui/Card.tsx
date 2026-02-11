import { ReactNode } from "react";

type CardPadding = "sm" | "md" | "lg";

interface CardProps {
  children: ReactNode;
  padding?: CardPadding;
  className?: string;
}

const paddingClasses: Record<CardPadding, string> = {
  sm: "p-4",
  md: "p-6",
  lg: "p-8",
};

export default function Card({
  children,
  padding = "md",
  className = "",
}: CardProps) {
  return (
    <div
      className={`
        bg-white rounded-2xl border border-sand
        shadow-[0_2px_8px_rgba(61,46,31,0.06)]
        ${paddingClasses[padding]}
        ${className}
      `}
    >
      {children}
    </div>
  );
}
