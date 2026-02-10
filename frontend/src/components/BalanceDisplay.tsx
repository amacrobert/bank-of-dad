interface BalanceDisplayProps {
  balanceCents: number;
  size?: "small" | "medium" | "large";
}

export default function BalanceDisplay({
  balanceCents,
  size = "medium",
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
  );
}
