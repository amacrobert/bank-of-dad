interface BalanceDisplayProps {
  balanceCents: number;
  size?: "small" | "medium" | "large";
  breakdown?: {
    availableCents: number;
    savedCents: number;
  };
}

export default function BalanceDisplay({
  balanceCents,
  size = "medium",
  breakdown,
}: BalanceDisplayProps) {
  const dollars = Math.floor(Math.abs(balanceCents) / 100);
  const cents = Math.abs(balanceCents) % 100;
  const isNegative = balanceCents < 0;
  const sign = isNegative ? "-" : "";

  const sizeConfig = {
    small: { dollar: "text-lg", cent: "text-xs", sign: "text-sm" },
    medium: { dollar: "text-2xl", cent: "text-sm", sign: "text-lg" },
    large: { dollar: "text-5xl", cent: "text-2xl", sign: "text-3xl" },
  };

  const s = sizeConfig[size];

  return (
    <div className="inline-flex flex-col items-center">
      <span className="inline-flex items-baseline font-bold text-forest tabular-nums">
        <span className={s.sign}>
          {sign}$
        </span>
        <span className={s.dollar}>
          {dollars.toLocaleString()}
        </span>
        <span className={`${s.cent} relative -top-[0.15em] ml-0.5 text-forest/70`}>
          {cents.toString().padStart(2, "0")}
        </span>
      </span>
      {breakdown && breakdown.savedCents > 0 && (
        <span className="text-sm text-bark-light mt-1">
          Available: ${(breakdown.availableCents / 100).toFixed(2)} · Saved: ${(breakdown.savedCents / 100).toFixed(2)}
        </span>
      )}
    </div>
  );
}
