interface BalanceDisplayProps {
  balanceCents: number;
  size?: "small" | "medium" | "large";
}

export default function BalanceDisplay({
  balanceCents,
  size = "medium",
}: BalanceDisplayProps) {
  const formatted = (balanceCents / 100).toFixed(2);

  const sizeClasses = {
    small: "balance-small",
    medium: "balance-medium",
    large: "balance-large",
  };

  return (
    <span className={`balance-display ${sizeClasses[size]}`}>${formatted}</span>
  );
}
